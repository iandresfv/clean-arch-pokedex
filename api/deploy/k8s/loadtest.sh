#!/usr/bin/env bash
#
# Demonstrates the rate limiter's behaviour against the running deployment.
#
# With one replica the configured limit holds exactly. With more than one it
# does not: a Kubernetes Service spreads a client's connections across pods,
# each pod keeps its own counters and therefore sees only a fraction of the
# traffic, and the effective limit becomes roughly N times what was configured.
# The client does not have to do anything clever — it simply receives more than
# it was promised, and the number moves on its own whenever the autoscaler acts.
#
# The load is generated from inside the cluster, and that detail is the whole
# point of the exercise: `kubectl port-forward` binds to one specific pod and
# bypasses Service load balancing entirely, so running this from the host would
# hide the very defect it exists to show.
#
# Measured on minikube with RATE_LIMIT_RPM=20, RATE_LIMIT_BURST=10, 30 requests:
#
#   1 replica,  in-process   allowed 10   throttled 20   <- the configured allowance
#   3 replicas, in-process   allowed 24   throttled  6   <- the defect, ~3x
#   3 replicas, Redis        allowed 10   throttled 20   <- fixed, matches 1 replica
#
# Usage:
#   ./loadtest.sh [namespace] [requests]

set -euo pipefail

NAMESPACE="${1:-pokedex-dev}"
REQUESTS="${2:-30}"

replicas=$(kubectl -n "${NAMESPACE}" get deploy pokedex-api \
    -o jsonpath='{.status.readyReplicas}' 2>/dev/null || echo '?')

echo "namespace      : ${NAMESPACE}"
echo "ready replicas : ${replicas}"
echo "requests       : ${REQUESTS}"
echo

kubectl -n "${NAMESPACE}" run "loadtest-$$" \
    --rm -i --restart=Never \
    --image=curlimages/curl:latest \
    --command -- sh -c "
allowed=0; throttled=0
for i in \$(seq 1 ${REQUESTS}); do
  # Connection: close forces a new TCP connection per request, which is what
  # makes the Service redistribute across pods. With keep-alive a single
  # connection sticks to one pod and the effect stays invisible.
  code=\$(curl -s -o /dev/null -w '%{http_code}' -H 'Connection: close' http://pokedex-api/api/v1/types)
  if [ \"\$code\" = '200' ]; then allowed=\$((allowed+1)); else throttled=\$((throttled+1)); fi
done
echo \"allowed (200)  : \$allowed\"
echo \"throttled (429): \$throttled\"
"

echo
if [ "${replicas}" != '?' ] && [ "${replicas}" -gt 1 ] 2>/dev/null; then
    echo "With ${replicas} replicas the in-process limiter admits roughly ${replicas}x the"
    echo "configured allowance: every pod keeps its own counters and none of them"
    echo "can see the whole picture."
    echo
    echo "Set REDIS_ENABLED=true and re-run: the counters move to a store all"
    echo "replicas share, and the limit holds."
fi
