package main_test

import (
	"os/exec"
	"testing"

	kusttest_test "sigs.k8s.io/kustomize/api/testutils/kusttest"
)

const fooInput = `
apiVersion: example.com/v1beta1
kind: MyPlugin
metadata:
  name: example-configmap-test
data:
  username: whatever
`

const fooExpect = `
kind: ConfigMap
apiVersion: v1
metadata:
  name: example-configmap-test
  annotations:
    kustomize.config.k8s.io/needs-hash: "false"
data:
  username: whatever
`

func TestMyPlugin(t *testing.T) {
	teardown := setup(t) // builds MyPlugin binary in source folder.
	defer teardown()

	th := kusttest_test.MakeEnhancedHarness(t).
		PrepExecPlugin("example.com", "v1beta1", "MyPlugin")
	defer th.Reset()

	t.Run("With foo", func(t *testing.T) {
		m := th.LoadAndRunGenerator(fooInput)
		th.AssertActualEqualsExpected(m, fooExpect)
	})
}

func setup(t *testing.T) func() {
	// Build plugin and place it in the test folder.
	cmd := exec.Command("go", "build", "-o", "MyPlugin", ".")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build plugin: %s", err)
	}
	return func() {
		// Placeholder for tearing down test.
	}
}
