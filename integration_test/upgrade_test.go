// Copyright 2019 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integrationtest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/krew/pkg/constants"
)

func TestKrewUpgrade_WithoutIndexInitialized(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()
	test.Krew("upgrade").RunOrFailOutput()
}

func TestKrewUpgrade(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().
		Krew("install", "--manifest", filepath.Join("testdata", validPlugin+constants.ManifestExtension)).
		RunOrFail()
	initialLocation := resolvePluginSymlink(test, validPlugin)

	test.Krew("upgrade").RunOrFail()
	eventualLocation := resolvePluginSymlink(test, validPlugin)

	if initialLocation == eventualLocation {
		t.Errorf("Expecting the plugin path to change but was the same.")
	}
}

func TestKrewUpgradeWhenPlatformNoLongerMatches(t *testing.T) {
	skipShort(t)

	test, cleanup := NewTest(t)
	defer cleanup()

	test.WithIndex().
		Krew("install", validPlugin).
		RunOrFail()

	test.WithEnv("KREW_OS", "somethingelse")

	// if upgrading 'all' plugins, must succeed
	out := string(test.Krew("upgrade", "--no-update-index").RunOrFailOutput())
	if !strings.Contains(out, "WARNING: Some plugins failed to upgrade") {
		t.Fatalf("upgrade all plugins output doesn't contain warnings about failed plugins:\n%s", out)
	}

	// if upgrading a specific plugin, it must fail, because no longer matching to a platform
	err := test.Krew("upgrade", validPlugin, "--no-update-index").Run()
	if err == nil {
		t.Fatal("expected failure when upgraded a specific plugin that no longer has a matching platform")
	}
}

func resolvePluginSymlink(test *ITest, plugin string) string {
	test.t.Helper()
	linkToPlugin, err := test.LookupExecutable("kubectl-" + plugin)
	if err != nil {
		test.t.Fatal(err)
	}

	realLocation, err := os.Readlink(linkToPlugin)
	if err != nil {
		test.t.Fatal(err)
	}

	return realLocation
}
