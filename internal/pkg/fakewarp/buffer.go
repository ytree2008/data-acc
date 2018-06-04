package fakewarp

import (
	"context"
	"fmt"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/keystoreregistry"
	"github.com/RSE-Cambridge/data-acc/internal/pkg/registry"
	"github.com/coreos/etcd/clientv3"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"
)

// Creates a persistent buffer.
// If it works, we return the name of the buffer, otherwise an error is returned
func DeleteBuffer(c CliContext, keystore keystoreregistry.Keystore) error {
	error := processDeleteBuffer(c.String("token"), keystore)
	return error
}

func processDeleteBuffer(bufferName string, keystore keystoreregistry.Keystore) error {
	r := keystoreregistry.NewBufferRegistry(keystore)
	// TODO: should do a get buffer before doing a delete
	buf := registry.Buffer{Name: bufferName}
	r.RemoveBuffer(buf)
	return nil
}

func CreatePerJobBuffer(c CliContext, keystore keystoreregistry.Keystore) error {
	error := processCreatePerJobBuffer(keystore, c.String("token"), c.Int("user"))
	return error
}

func processCreatePerJobBuffer(keystore keystoreregistry.Keystore, token string, user int) error {
	r := keystoreregistry.NewBufferRegistry(keystore)
	// TODO: lots more validation needed to ensure valid key, etc
	buf := registry.Buffer{Name: token, Owner: fmt.Sprintf("%d", user)}
	r.AddBuffer(buf)
	return nil
}

func getBricks(cli *clientv3.Client, prefix string) map[string]map[string]registry.Brick {
	allBricks := make(map[string]map[string]registry.Brick)
	getResponse, err := cli.Get(context.Background(), prefix, clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, keyValue := range getResponse.Kvs {
		rawKey := fmt.Sprintf("%s", keyValue.Key) // e.g. /bricks/present/1aff0f8468ee/nvme7n1
		key := strings.Split(rawKey, "/")
		brick := registry.Brick{Name: key[4], Hostname: key[3]}
		_, ok := allBricks[brick.Hostname]
		if !ok {
			allBricks[brick.Hostname] = make(map[string]registry.Brick)
		}
		allBricks[brick.Hostname][brick.Name] = brick
	}
	return allBricks
}

func getAvailableBricks(cli *clientv3.Client) map[string][]registry.Brick {
	allBricks := getBricks(cli, "/bricks/present/")
	inUseBricks := getBricks(cli, "/bricks/inuse/")

	aliveHosts := make(map[string]string)
	getHostsResponse, err := cli.Get(context.Background(), "/bufferhost/alive/", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}
	for _, keyValue := range getHostsResponse.Kvs {
		rawKey := fmt.Sprintf("%s", keyValue.Key)
		key := strings.Split(rawKey, "/") // e.g. /bufferhost/alive/afe30ea9f27e
		host := key[3]
		aliveHosts[host] = rawKey
	}

	availableBricks := make(map[string][]registry.Brick)

	for host, allHostBricks := range allBricks {
		aliveHost, ok := aliveHosts[host]
		if !ok || aliveHost == "" {
			continue
		}
		inuseHostBricks := inUseBricks[host]

		availableBricks[host] = []registry.Brick{}

		for _, brick := range allHostBricks {
			inuse := false
			for _, inUseBrick := range inuseHostBricks {
				if inUseBrick.Name == brick.Name {
					inuse = true
					break
				}
			}
			if !inuse {
				availableBricks[host] = append(availableBricks[host], brick)
			}
		}
	}
	return availableBricks
}

func AddFakeBufferAndBricks(keystore keystoreregistry.Keystore, cli *clientv3.Client) registry.Buffer {
	log.Println("Add fakebuffer and match to bricks")
	bufferRegistry := keystoreregistry.NewBufferRegistry(keystore)
	availableBricks := getAvailableBricks(cli)
	var chosenBricks []registry.Brick

	// pick some of the available bricks
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s) // initialize local pseudorandom generator
	requestedBricks := 2

	var hosts []string
	for key := range availableBricks {
		hosts = append(hosts, key)
	}

	randomWalk := rand.Perm(len(availableBricks))
	for _, i := range randomWalk {
		hostBricks := availableBricks[hosts[i]]
		candidateBrick := hostBricks[r.Intn(len(hostBricks))]

		goodCandidate := true
		for _, brick := range chosenBricks {
			if brick == candidateBrick {
				goodCandidate = false
				break
			}
			if brick.Hostname == candidateBrick.Hostname {
				goodCandidate = false
				break
			}
		}
		if goodCandidate {
			chosenBricks = append(chosenBricks, candidateBrick)
		}

		if len(chosenBricks) >= requestedBricks {
			break
		}
	}
	// TODO: check we have enough bricks?

	bufferName, _ := os.Hostname()
	log.Printf("For buffer %s selected following bricks: %s\n", bufferName, chosenBricks)

	// TODO: should be done in a single transaction, and retry if clash
	for i, brick := range chosenBricks {
		chosenKey := fmt.Sprintf("/bricks/inuse/%s/%s", brick.Hostname, brick.Name)
		keystore.AtomicAdd(chosenKey, fmt.Sprintf("%s:%d", bufferName, i))
	}

	buffer := registry.Buffer{Name: bufferName, Bricks: chosenBricks}
	bufferRegistry.AddBuffer(buffer)
	return buffer
}