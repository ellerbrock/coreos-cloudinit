package cloudinit

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"
	"testing"
)

func TestEtcdEnvironment(t *testing.T) {
	cfg := make(EtcdEnvironment, 0)
	cfg["discovery_url"] = "http://disco.example.com/foobar"
	cfg["peer-bind-addr"] = "127.0.0.1:7002"

	env := cfg.String()
	expect := `[Service]
Environment="ETCD_DISCOVERY_URL=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`

	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestEtcdEnvironmentReplacement(t *testing.T) {
	os.Clearenv()
	os.Setenv("COREOS_PUBLIC_IPV4", "203.0.113.29")
	os.Setenv("COREOS_PRIVATE_IPV4", "192.0.2.13")

	cfg := make(EtcdEnvironment, 0)
	cfg["bind-addr"] = "$public_ipv4:4001"
	cfg["peer-bind-addr"] = "$private_ipv4:7001"

	env := cfg.String()
	expect := `[Service]
Environment="ETCD_BIND_ADDR=203.0.113.29:4001"
Environment="ETCD_PEER_BIND_ADDR=192.0.2.13:7001"
`
	if env != expect {
		t.Errorf("Generated environment:\n%s\nExpected environment:\n%s", env, expect)
	}
}

func TestEtcdEnvironmentWrittenToDisk(t *testing.T) {
	ec := EtcdEnvironment{
		"discovery_url": "http://disco.example.com/foobar",
		"peer-bind-addr": "127.0.0.1:7002",
	}
	dir, err := ioutil.TempDir(os.TempDir(), "coreos-cloudinit-")
	if err != nil {
		t.Fatalf("Unable to create tempdir: %v", err)
	}
	defer syscall.Rmdir(dir)

	if err := WriteEtcdEnvironment(dir, ec); err != nil {
		t.Fatalf("Processing of EtcdEnvironment failed: %v", err)
	}

	fullPath := path.Join(dir, "etc", "systemd", "system", "etcd.service.d", "20-cloudinit.conf")

	fi, err := os.Stat(fullPath)
	if err != nil {
		t.Fatalf("Unable to stat file: %v", err)
	}

	if fi.Mode() != os.FileMode(0644) {
		t.Errorf("File has incorrect mode: %v", fi.Mode())
	}

	contents, err := ioutil.ReadFile(fullPath)
	if err != nil {
		t.Fatalf("Unable to read expected file: %v", err)
	}

	expect := `[Service]
Environment="ETCD_DISCOVERY_URL=http://disco.example.com/foobar"
Environment="ETCD_PEER_BIND_ADDR=127.0.0.1:7002"
`
	if string(contents) != expect {
		t.Fatalf("File has incorrect contents")
	}
}

func rmdir(path string) error {
    cmd := exec.Command("rm", "-rf", path)
    return cmd.Run()
}
