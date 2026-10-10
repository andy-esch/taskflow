// Optional development check against an installed Renovate, not a CLI/runtime dependency.
// Usage: node internal/tools/releasevalidate/renovate_policy_check.mjs /path/to/node_modules/renovate
// Internal Renovate imports are intentional: use its extraction/matching, not a local imitation.
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath, pathToFileURL } from 'node:url';

const root = fileURLToPath(new URL('../../../', import.meta.url));
const renovateRoot = process.argv[2];
assert(renovateRoot, 'Pass the installed Renovate package directory');
const load = (file) => import(pathToFileURL(path.resolve(renovateRoot, 'dist', file)));
const { GlobalConfig } = await load('config/global.js');
GlobalConfig.set({ localDir: root });
const { init } = await load('logger/index.js');
await init();
const { applyPackageRules } = await load('util/package-rules/index.js');
const { extractPackageFile: extractGo } = await load('modules/manager/gomod/extract.js');
const { extractPackageFile: extractActions } = await load('modules/manager/github-actions/extract.js');
const { extractPackageFile: extractDocker } = await load('modules/manager/dockerfile/extract.js');
const { extractPackageFile: extractRegex } = await load('modules/manager/custom/regex/index.js');
const config = JSON.parse(await readFile(path.join(root, 'renovate.json'), 'utf8'));
const file = async (name) => readFile(path.join(root, name), 'utf8');
const containerPath = 'build/release-validation/Containerfile';
const container = await file(containerPath);
const go = extractGo(await file('go.mod')).deps.find((dep) => dep.depType === 'golang');
assert(go && go.depName === 'go', 'go.mod minimum must be extracted');
const actions = (await extractActions(await file('.github/workflows/ci.yml'), '.github/workflows/ci.yml', {})).deps;
const workflowGo = actions.find((dep) => dep.depName === 'go');
const workflowLint = actions.find((dep) => dep.depName === 'golangci/golangci-lint');
assert(workflowGo && !workflowGo.skipReason, 'setup-go selector must be extracted');
assert(workflowLint && !workflowLint.skipReason, 'CI linter selector must be extracted');
const docker = extractDocker(container, containerPath, {}).deps.find((dep) => dep.depName === 'golang');
assert(docker && !docker.skipReason, 'Containerfile Go image must be extracted');
const custom = config.customManagers.find((manager) => manager.customType === 'regex');
assert(custom, 'Container linter needs a regex manager');
assert(custom.managerFilePatterns.some((pattern) => new RegExp(pattern.slice(1, -1)).test(containerPath)), 'Regex manager must discover Containerfile');
const lint = extractRegex(container, containerPath, custom)?.deps.find((dep) => dep.depName === 'golangci/golangci-lint');
assert(lint && !lint.skipReason, 'Container linter ARG must be extracted');

async function policy(dep, manager, updateType) {
  return applyPackageRules({ ...config, ...dep, manager, updateType }, 'update');
}
for (const [dep, manager] of [[workflowGo, 'github-actions'], [docker, 'dockerfile']]) {
  const upgrade = await policy(dep, manager, 'minor');
  assert.equal(upgrade.groupName, 'go version');
  assert.equal(upgrade.dependencyDashboardApproval, true);
  assert.equal(upgrade.separateMinorPatch, true);
  for (const type of ['patch', 'digest', 'pinDigest']) {
    const patch = await policy(dep, manager, type);
    assert.equal(patch.groupName, 'go toolchain patches');
    assert.equal(patch.dependencyDashboardApproval, false);
    assert.deepEqual(patch.schedule, ['at any time']);
    assert.equal(patch.updateNotScheduled, true);
    assert.equal(patch.minimumReleaseAge, '0 days');
    assert.equal(patch.automerge, false);
  }
}
for (const type of ['minor', 'patch']) {
  const minimum = await policy(go, 'gomod', type);
  assert.equal(minimum.groupName, 'go version');
  assert.equal(minimum.dependencyDashboardApproval, true);
  assert.equal(minimum.rangeStrategy, 'bump');
}
for (const [dep, manager] of [[workflowLint, 'github-actions'], [lint, 'custom.regex']]) {
  const result = await policy(dep, manager, 'minor');
  assert.equal(result.groupName, 'go lint tooling');
  assert.equal(result.dependencyDashboardApproval, true);
}
const checkout = actions.find((dep) => dep.depName === 'actions/checkout');
assert(checkout, 'Unrelated action control must exist');
assert.equal((await policy(checkout, 'github-actions', 'minor')).groupName, 'github actions');
console.log(`Renovate ${(JSON.parse(await readFile(path.resolve(renovateRoot, 'package.json'), 'utf8'))).version}: extraction and toolchain policy checks passed`);
