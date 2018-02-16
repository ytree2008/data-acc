# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License. You may obtain
# a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and limitations
# under the License.

import json
import time

from cliff.command import Command

from burstbuffer import fakewarp_facade


def _output_as_json(cmd, output):
    json.dump(output, cmd.app.stdout, sort_keys=True, indent=4)


class Pools(Command):
    """Output burst buffer pools"""

    def take_action(self, parsed_args):
        _output_as_json(self, fakewarp_facade.get_pools())


class ShowInstances(Command):
    """Show burst buffers instances"""

    def take_action(self, parsed_args):
        _output_as_json(self, fakewarp_facade.get_instances())


class ShowSessions(Command):
    """Show burst buffers sessions"""

    def take_action(self, parsed_args):
        _output_as_json(self, fakewarp_facade.get_sessions())


class Teardown(Command):
    """Start the teardown of the given burst buffer"""

    def get_parser(self, prog_name):
        parser = super(Teardown, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--hurry', action="store_true", default=False)
        return parser

    def take_action(self, parsed_args):
        print(parsed_args.job_id)
        print(parsed_args.buffer_script)
        print(parsed_args.hurry)


class JobProcess(Command):
    """Initial call when job is run to parse buffer script."""

    def get_parser(self, prog_name):
        parser = super(JobProcess, self).get_parser(prog_name)
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        job_config_line = None
        with open(parsed_args.buffer_script) as f:
            for line in f:
                if line.startswith("#DW jobdw"):
                    job_config_line = line
                    break

        config = job_config_line.strip("#DW jobdw ")
        print(config)
        # this validates the buffer script, next step is calling "setup"
        # once there is enough available bust buffer space


class Setup(Command):
    """Create the burst buffer, ready to start the data stage_in"""

    def get_parser(self, prog_name):
        parser = super(Setup, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--caller', type=str,
                            help="Caller, i.e. SLURM")
        parser.add_argument('--user', type=int,
                            help="User id, i.e. 1001")
        parser.add_argument('--groupid', type=int,
                            help="Group id, i.e. 1001")
        parser.add_argument('--capacity', type=str,
                            help="The pool and capacity, i.e. dwcache:1GiB")
        return parser

    def take_action(self, parsed_args):
        # this should add the burst buffer in the DB, so real_size works
        print(parsed_args.job_id)
        print(parsed_args.buffer_script)
        print(parsed_args.capacity)
        print("pool: %s, capacity: %s" % tuple(
            parsed_args.capacity.split(":")))


class RealSize(Command):
    """Report actual size of burst buffer, rounded up for granularity"""

    def get_parser(self, prog_name):
        parser = super(RealSize, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        return parser

    def take_action(self, parsed_args):
        fake_size = {
            "token": parsed_args.job_id,
            "capacity": 17592186044416,
            "units": "bytes"
        }
        _output_as_json(self, fake_size)


class DataIn(Command):
    """Start copy of data into the burst buffer"""

    def get_parser(self, prog_name):
        parser = super(DataIn, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        # although I think this is async...
        time.sleep(10)
        # "No matching session" if there is no matching job found


class Paths(Command):
    """Output paths to share with job in environment file"""

    def get_parser(self, prog_name):
        parser = super(Paths, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--pathfile', type=str,
                            help="Path to write out environment variables.")
        return parser

    def take_action(self, parsed_args):
        # Test with: sbatch --bbf=buffer.txt --wrap="echo $DW_PATH_TEST"
        with open(parsed_args.pathfile, "w") as f:
            f.write("DW_PATH_TEST=/tmp/dw")


class PreRun(Command):
    """Do setup on compute nodes prior to running job"""
    def get_parser(self, prog_name):
        parser = super(PreRun, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        parser.add_argument('--nodehostnamefile', type=str,
                            help="Path file containing chosen nodes.")
        return parser

    def take_action(self, parsed_args):
        with open(parsed_args.nodehostnamefile) as f:
            print("".join(f.readlines()))


class PostRun(Command):
    """Do post run cleanup, before data stage out."""
    def get_parser(self, prog_name):
        parser = super(PostRun, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        print(parsed_args.job_id)


class DataOut(Command):
    """Copy data out of burst buffer."""
    def get_parser(self, prog_name):
        parser = super(DataOut, self).get_parser(prog_name)
        parser.add_argument('--token', type=str, dest="job_id",
                            help="Job ID")
        parser.add_argument('--job', type=str, dest="buffer_script",
                            help="Path to burst buffer script file.")
        return parser

    def take_action(self, parsed_args):
        print(parsed_args.job_id)


class ShowConfigurations(Command):
    """Fake command to keep Slurm plugin happy."""
    def take_action(self, parsed_args):
        # slurm just ignores the output
        _output_as_json(self, {"configurations": []})


class CreatePersistent(Command):
    """Create a persistent burst buffer. Teardown is used to delete buffer."""
    def get_parser(self, prog_name):
        parser = super(CreatePersistent, self).get_parser(prog_name)
        parser.add_argument('--token', '-t', type=str, dest="name",
                            help="Name of the persistent buffer.")
        parser.add_argument('--caller', '-c', type=str,
                            help="Caller of the script, e.g. CLI or SLURM.")
        parser.add_argument('--capacity', '-C', type=str,
                            help="Capacity in the form <pool>:<no of bytes>.")
        parser.add_argument('--user', '-u', type=str,
                            help="User that owns the buffer.")
        parser.add_argument('--group', '-g', type=str,
                            help="Currently ignored, not used by Slurm.")
        parser.add_argument('--access', '-a', type=str,
                            help="Access mode, e.g. striped.")
        parser.add_argument('--type', '-T', type=str,
                            help="Buffer type, e.g. scratch.")
        return parser

    def take_action(self, parsed_args):
        print(parsed_args.name)
        pool_name, capacity_bytes = parsed_args.capacity.split(":")
        fakewarp_facade.add_persistent_buffer(
            parsed_args.name, parsed_args.caller,
            pool_name, capacity_bytes,
            parsed_args.user, parsed_args.access, parsed_args.type)
