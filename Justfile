refresh-fixture-inputs:
    curl 'https://api.openshift.com/api/upgrades_info/graph?channel=candidate-4.20' > pkg/fauxinnati/testdata/zz_fixture_TestGraph_JSONcandidate_4.20.json.input
