package cutlass

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func InternetTraffic(bp_dir, fixture_path, buildpack_path string, envs []string) ([]string, error) {
	network_command := "(sudo tcpdump -n -i eth0 not udp port 53 and not udp port 1900 and not udp port 5353 and ip -t -Uw /tmp/dumplog &) && /buildpack/bin/detect /tmp/staged && /buildpack/bin/compile /tmp/staged /tmp/cache && /buildpack/bin/release /tmp/staged /tmp/cache && pkill tcpdump; tcpdump -nr /tmp/dumplog | sed -e 's/^/internet traffic: /' 2>&1 || true"

	output, err := executeDockerFile(bp_dir, fixture_path, buildpack_path, envs, network_command)
	if err != nil {
		return nil, err
	}

	var out []string
	for _, line := range strings.Split(output, "\n") {
		if idx := strings.Index(line, "internet traffic: "); idx >= 0 && idx < 10 {
			out = append(out, line[(idx+18):])
		}
	}

	return out, nil
}

func UniqueDestination(traffic []string, destination string) error {
	re := regexp.MustCompile("^[\\d\\.:]+ IP ([\\d\\.]+) > ([\\d\\.]+):")
	for _, line := range traffic {
		m := re.FindStringSubmatch(line)
		if len(m) != 3 || (m[1] != destination && m[2] != destination) {
			return fmt.Errorf("Outgoing traffic: %s", line)
		}
	}
	return nil
}

func executeDockerFile(bp_dir, fixture_path, buildpack_path string, envs []string, network_command string) (string, error) {
	var err error
	buildpack_path, err = filepath.Rel(bp_dir, buildpack_path)

	docker_image_name := "internet_traffic_test"

	// docker_env_vars += get_app_env_vars(fixture_path)
	dockerfile_contents := dockerfile(fixture_path, buildpack_path, envs, network_command)

	dockerfile_name := fmt.Sprintf("itf.Dockerfile.%v", rand.Int())
	err = ioutil.WriteFile(filepath.Join(bp_dir, dockerfile_name), []byte(dockerfile_contents), 0755)
	if err != nil {
		return "", err
	}
	defer os.Remove(filepath.Join(bp_dir, dockerfile_name))
	defer exec.Command("docker", "rmi", "-f", docker_image_name).Output()

	cmd := exec.Command("docker", "build", "--rm", "--no-cache", "-t", docker_image_name, "-f", dockerfile_name, ".")
	cmd.Dir = bp_dir
	cmd.Stderr = DefaultStdoutStderr
	output, err := cmd.Output()

	return string(output), err
}

func dockerfile(fixture_path, buildpack_path string, envs []string, network_command string) string {
	out := "FROM cloudfoundry/cflinuxfs2\n" +
		"ENV CF_STACK cflinuxfs2\n" +
		"ENV VCAP_APPLICATION {}\n"
	for _, env := range envs {
		out = out + "ENV " + env + "\n"
	}
	out = out +
		"ADD " + fixture_path + " /tmp/staged/\n" +
		"ADD " + buildpack_path + " /tmp/\n" +
		"RUN mkdir -p /buildpack\n" +
		"RUN mkdir -p /tmp/cache\n" +
		"RUN unzip -o /tmp/" + filepath.Base(buildpack_path) + " -d /buildpack\n" +
		"# HACK around https://github.com/dotcloud/docker/issues/5490\n" +
		"RUN mv /usr/sbin/tcpdump /usr/bin/tcpdump\n" +
		"RUN " + network_command + "\n"
	return out
}
