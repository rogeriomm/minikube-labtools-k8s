#!/usr/bin/env python3

# https://newbedev.com/6-ways-to-call-external-command-in-python

import os
import sys
import docker
from git import Repo
import getpass

from rich.console import Console
from rich.theme import Theme
from rich.traceback import install
from rich.progress import Progress

_PY38_MIN = sys.version_info[:2] >= (3, 10)

if not _PY38_MIN:
    raise SystemExit('ERROR: make requires a minimum of Python3 version 3.10 or later. Current version: %s' % ''.join(
        sys.version.splitlines()))


class DockerBuildComponent:
    """ Docker build component
    Image naming conventions: REGISTRY[:PORT]/USER/REPO:TAG

    REPO = {prefix}-{name}-{version}:{git branch}
    """

    def __init__(self, name: str, username=None, version: str = '', parm={}, prefix='', depends=[], tobuild=True):
        self.name = name
        self.username = username
        self.version = version
        self.parm = parm
        self.prefix = prefix
        self.tobuild = tobuild

        self.__repo = Repo('.')
        self.__dir = os.getcwd()
        self.__branch = self.__repo.active_branch.name

    def show(self):
        console.print(
            f"   name: \"{self.name}\", prefix: \"{self.prefix}\", version: \"{self.version}\", parm: {self.parm}, build: {self.tobuild} ")

    def get_docker_repo(self) -> str:
        repo = f"{registry_name}/" if registry_name is not None else ""
        repo += f"{self.username}/" if self.username is not None else ""
        return repo

    def get_docker_name(self) -> str:
        tag = None
        if self.prefix == '' and self.version == '':
            tag = f"{self.name}"
        elif self.prefix == '' and self.version != '':
            tag = f"{self.name}-{self.version}"
        elif self.prefix != '' and self.version != '':
            tag = f"{self.prefix}-{self.name}-{self.version}"
        elif self.prefix != '' and self.version == '':
            tag = f"{self.prefix}-{self.name}"

        tag = f"{tag}:{self.__branch}"

        repo = self.get_docker_repo()
        if repo != "":
            tag = f"{repo}{tag}"

        return tag

    def minikube_push(self):
        console.print(f"Minikube push: {self.get_docker_name()}")
        cmd = "minikube -p cluster2 image push " + self.get_docker_name()
        console.print(cmd)

    def docker_push(self):
        console.print(f"Docker push: {self.get_docker_name()}")

        ret = True
        chunk_progress = {}

        with Progress(auto_refresh=True, refresh_per_second=3, expand=True, transient=False) as progress:
            for chunk in cli.push(repository=self.get_docker_name(), stream=True, decode=True):
                if 'progressDetail' in chunk:
                    if 'status' in chunk:
                        if chunk["status"] == "Pushing":
                            if 'total' in chunk["progressDetail"]:
                                progress.update(chunk_progress[chunk["id"]],
                                                description="[yellow]Pushing " + chunk["id"],
                                                total=chunk["progressDetail"]["total"],
                                                completed=chunk["progressDetail"]["current"])
                            else:
                                continue
                        elif chunk["status"] == "Preparing":
                            chunk_progress[chunk["id"]] = progress.add_task(
                                description="[blue]Preparing " + chunk["id"],
                                visible=True, start=True)
                        elif chunk["status"] == "Waiting":
                            progress.update(chunk_progress[chunk["id"]], description="[red]Waiting " + chunk["id"])
                        elif 'id' in chunk:
                            progress.update(chunk_progress[chunk["id"]],
                                            description="[green]" + chunk["status"] + " " + chunk["id"])
                    else:
                        progress.console.print(chunk)
                elif 'status' in chunk:
                    progress.console.print(chunk["status"])
                elif 'errorDetail' in chunk:
                    progress.console.print(chunk['errorDetail']['message'])
                else:
                    progress.console.print(f":bug: {chunk}")
        return ret

    # https://docker-py.readthedocs.io/en/stable/api.html#module-docker.api.build
    def build(self, nocache: bool) -> bool:

        ret = True

        args = {"TAG": self.__branch, "REPO": self.get_docker_repo(), "VERSION": self.version}
        args.update(self.parm)
        console.print(args)

        tag = self.get_docker_name()

        console.print(f"Building {tag} {args}", style="bold red on black")

        streamer = cli.build(path=f"./{self.name}", tag=tag,
                             nocache=nocache, rm=False, buildargs=args,
                             forcerm=True,
                             decode=True)

        for chunk in streamer:
            if 'stream' in chunk:
                for line in chunk['stream'].splitlines():
                    print(line)
            elif 'status' in chunk:
                print(f"{chunk}")
            elif 'message' in chunk:
                for line in chunk['message'].splitlines():
                    print(f"{line}")
                ret = False
            elif 'aux' in chunk:
                pass
            elif 'errorDetail' in chunk:
                console.print(f"{chunk['errorDetail']}", style="error")
                ret = False
            elif "\n" in chunk:
                print("")
            else:
                console.print(f":bug: Unknown chunk {chunk}", style="error")

        return ret


class DirComponent:
    """ Directory """

    def __init__(self, name="", dirs=[]):
        self.name = name
        self.dirs = dirs
        self.scanned = False

    def show(self):
        console.print(f"   name: {self.name}: dirs: {self.dirs}")


def add_mk(pkg: [], name: str):
    global flprj
    flprj.add_pkg(name, pkg)


def dir_mk(d: [], name: str):
    global flprj
    flprj.add_dir(name, d)


class BuildMk:
    """ build.mk file """

    def __init__(self):
        self.__hasPrjFile = False
        self.__pkgs = []
        self.__dirs = []
        self.__dir = os.getcwd()

    def has_prj_file(self) -> bool:
        return self.__hasPrjFile

    def get_repos_name(self):
        return os.path.basename(self.__dir)

    def get_pkgs(self) -> []:
        return self.__pkgs

    def get_dirs(self) -> []:
        return self.__dirs

    def get_prj_dir(self) -> str:
        return self.__dir

    def add_pkg(self, name='', p=[]):
        if name != '' and len(p) > 0:
            self.__pkgs.append(p)

    def add_dir(self, name='', d=[]):
        if name != '' and len(d) > 0:
            self.__dirs.append(DirComponent(name, d))

    def scan(self):
        global flprj
        os.chdir(self.__dir)
        try:
            flprj = self
            self.__pkgs = []
            self.__dirs = []
            with open('build.mk', mode='r', encoding='utf8') as f:
                eval(f.read(),
                     {'__builtins__': None}, {'Docker': DockerBuildComponent, 'Prj': add_mk, 'Dir': dir_mk})
                f.close()
                self.__hasPrjFile = True
        except IOError:
            self.__hasPrjFile = False
        finally:
            flprj = None

    def build(self, nocache=False):
        os.chdir(self.__dir)
        for p in self.__pkgs:
            if type(p) is list:
                for c in p:
                    if c.tobuild and not c.build(nocache):
                        console.print(f"Building \"{c.get_docker_name()}\" failed.", style="error")
                        break  # Build failed
            else:
                break  # Failed

    def minikube_push(self):
        os.chdir(self.__dir)
        for p in self.__pkgs:
            if type(p) is list:
                for c in p:
                    c.minikube_push()

    def docker_push(self):
        os.chdir(self.__dir)
        for p in self.__pkgs:
            if type(p) is list:
                for c in p:
                    if not c.docker_push():
                        console.print(f"Pushing \"{c.get_docker_name()}\" failed.", style="error")
                        break  # Push failed
            else:
                break  # Failed

    def show(self):
        console.print(f"➡️  {self.__dir}")
        if self.__hasPrjFile:
            for p in self.__pkgs:
                if type(p) is list:
                    for c in p:
                        c.show()
            if len(self.__dir) > 0:
                for p in self.__dirs:
                    if type(p) is DirComponent:
                        p.show()


class AllBuildMk:
    """ All projects """

    def __init__(self):
        self.__prjs = []
        self.__dir = os.getcwd()

    def scan(self):
        """ Scan recursively """
        p = BuildMk()
        p.scan()
        self.__prjs.append(p)
        if p.has_prj_file():
            for d in p.get_dirs():
                cur_dir = os.getcwd()  # save current directory
                for i in d.dirs:
                    os.chdir(i)
                    self.scan()
                    os.chdir(cur_dir)

    def show(self):
        for p in self.__prjs:
            if type(p) is BuildMk:
                p.show()

    def build(self):
        for p in self.__prjs:
            if type(p) is BuildMk:
                p.build()
            else:
                break

    def command(self, cd: []) -> bool:
        cur_dir = os.getcwd()

        match cd:
            case ["show"]:
                self.show()
                return True

        try:
            for p in self.__prjs:
                console.print(f"🔄 {p.get_repos_name()}")
                os.chdir(p.get_prj_dir())
                match cd:
                    case ["build"]:
                        p.build(nocache=False)

                    case ["rebuild"]:
                        p.build(nocache=True)

                    case ["docker", "push"]:
                        p.docker_push()

                    case ["minikube", "push"]:
                        p.minikube_push()

                    case ["add", "origin"]:
                        cmd = f"git remote add origin git@github.com:{github_username}/{p.get_repos_name()}.git"
                        os.system(cmd)

                    case ["remove", "origin"]:
                        cmd = f"git remote remove origin"
                        os.system(cmd)

                    case ["git", "push", "all"]:
                        cmd = f"git push --all origin"
                        os.system(cmd)

                    case ["delete", "repos", "github"]:
                        cmd = f"gh repo-delete {github_username}/{p.get_repos_name()}"
                        os.system(cmd)

                    case ["create", "repos", where]:
                        match where:
                            case "github":
                                cmd = f"gh repo create {github_username}/{p.get_repos_name()} --confirm --public " + \
                                      f"--description \"{p.get_repos_name()}\""
                                os.system(cmd)
                            case "bitbucket":
                                cmd = f""
                                os.system(cmd)

                    case ["ps"]:
                        if os.path.isfile("docker-compose.yml"):
                            cmd = f"export TAG=$(git rev-parse --abbrev-ref HEAD) ; docker-compose -f " \
                                  f"docker-compose.yml ps"
                            os.system(cmd)

                    case ["start"]:
                        if os.path.isfile("docker-compose.yml"):
                            cmd = f"export TAG=$(git rev-parse --abbrev-ref HEAD) ; docker-compose -f " \
                                  f"docker-compose.yml up -d"
                            os.system(cmd)

                    case ["stop"]:
                        if os.path.isfile("docker-compose.yml"):
                            cmd = f"export TAG=$(git rev-parse --abbrev-ref HEAD) ; docker-compose -f " \
                                  f"docker-compose.yml down"
                            os.system(cmd)

                    case ["shell", name]:
                        if os.path.isfile("docker-compose.yml"):
                            cmd = f"export TAG=$(git rev-parse --abbrev-ref HEAD) ; docker-compose -f " \
                                  f"docker-compose.yml exec {name} /bin/bash"
                            os.system(cmd)
                            return True

                    case _:
                        return False

        finally:
            os.chdir(cur_dir)

        return True


def init_docker():
    global cli
    #dc = docker.from_env()

    #tls_config = docker.tls.TLSConfig(
    #    ca_cert=dc.api.verify,
    #    client_cert=dc.api.cert,
    #    verify=True)

    # https://www.programcreek.com/python/?code=picoCTF%2FpicoCTF%2FpicoCTF-master%2FpicoCTF-shell%2Fhacksport%2Fdocker.py: Search "tls_config"
    #cli = docker.APIClient(base_url=dc.api.base_url, tls=tls_config)
    cli = docker.APIClient()


def init():
    global so_username
    global registry_name
    global console
    global github_username

    # Install Rich
    install()

    flprj: BuildMk

    # Docker framework
    dc = None
    cli = None

    custom_theme = Theme({"success": "green", "error": "bold red"})
    console = Console(theme=custom_theme)

    so_username = getpass.getuser()
    github_username = "rogeriomm"
    registry_name = "jfrog.worldl.xpt/lab"

    init_docker()


def main():
    init()

    p = AllBuildMk()
    p.scan()

    argv = sys.argv[1:]
    if len(argv) == 0:
        return

    if not p.command(argv):
        console.print("Invalid command")


if __name__ == '__main__':
    main()
