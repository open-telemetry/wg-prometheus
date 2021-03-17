# collector-load-balancer

NOTE: this is WIP prototype based on and internal prototype for demo purpose.

## Usage

```bash
# init
make tools
./hack/kind_create_cluster

# build and deploy
make clb-docker clbotelcol-docker
make deploy

# schedule w/o auto scale
kubectl apply -f config/samples/otel_v1_collectorloadbalancer.yaml
kubectl scale --replicas=5 collectorloadbalancers/collectorloadbalancer-sample
kubectl scale --replicas=1 collectorloadbalancers/collectorloadbalancer-sample

# scale w/ target based auto scale
kubectl apply -f config/samples/otel_v1_collectorloadbalancer_scale.yaml
```

## Read the code

The main idea is when you change replicas, you need to reschedule targets. You can use scheduler (i.e. reshard) without
any auto scale and manually scale it via kubectl. Its builtin auto scale is naive for showing how to auto scale w/o
external autoscaler(e.g. hpa).

```
k apply -f crd.yaml

watch {
  crd + sts // you can't even tell which is which thanks to controller runtime ...
    sync <- newSpec/Status
  targets
    relabel, sync <- discoverdTargets
  timer
    sync <- gc
}

sync {
  // Scale
  if crd.Autoscale {
     expectedState := scaler.ExpectedReplicas(discoverdTargets) // internal scaler
  } else {
     expectedState := crd.Replicas // changed by user or external auto sacler
  }
  k8s.scaleSts(expectedState)
  
  // Scedule
  newState := sceduler.Schedule(existingState, expcetdState, targets) {
    curr := existingState.DeepCopy()
    if replicaChanged {
      // Reshard, ignore existing schedule
      curr.ClearTargets()
    } else {
      // Keep existing state as much as possible
      targets := diff(curr, targets)
    }
    scheduleTargets(curr, targets)
  }
  
  // Update collectors
  dispatcher.Send(collectors, newState)
}
```

- config
    - [manual scale](config/samples/otel_v1_collectorloadbalancer.yaml)
    - [auto scale](config/samples/otel_v1_collectorloadbalancer_scale.yaml)
- server
    - [api](api/v1/collectorloadbalancer_types.go) defines crd
    - [controller](controllers/collectorloadbalancer_controller.go) entry point for watching k8s resources
    - [instance](loadbalancer/instance.go) the actual sync logic, starts promsd, calls scheduler and scaler, send
      schedule result to configmanager
    - [scheduler](loadbalancer/scheduler.go) a naive greedy scheduler (i.e. just a priority queue)
    - [scaler](loadbalancer/scaler.go) a naive scaler based on expected targets per instance
- client
    - [cmd/clbotelcol](cmd/clbotelcol) the otel distro w/ configmanager
    - [configmanager](configmanager) the extension that writes prometheus file sd
- prometheus
    - [promsd](loadbalancer/promsd) run prometheus sd manager on server, do relabel to drop targets before sending to
      scheduler
    - [configpaerser](loadbalancer/configparser) extract and replace prometheus config, used in both server and client (
      indirectly)
- [proto](proto/src) (most of them are not implemeted except configmanager, which is broken see known issue...)

Please ignore the (empty) CRD for load generator, it was meant for testing HPA based on cpu, mem metrics etc and see how
many targets is required to overload a fixed number of collector (asked by Richard Anton).

## Known issues

Major

- it does not work with push based receiver because there is no consistent endpoint for sdk (in applications) to push
  metrics/trace to.
- only sts is supported, ds can be supported but maybe it's just easier for ds pods to run leader election

Minor

- hpa is not working, but should be trivial by applying hpa on sts deployed by controller, the semantic is tricky when
  apply hpa on crd itself directly.
- prometheus target key is not that unique, many targets from k8s sd has same `__address__` but different labels
- can't send targets list to collector via grpc, both response and error are nil XD (It was working in old version)
- update crd is not fully supported, it does not redeploy collector/load new sd config (like the old version)
- crd status is not updated, status.Replica is always 0 ...

## Authors

- pingleig@ run kubebuilder and kubectl
- zmengyi@ scheduler and scaler
- xiami@ collector extension

## Acknowledgement

- https://github.com/tkestack/kvass and https://github.com/gaocegege Our design is very similar to KVASS, except we
  focus on scraping part and can support other workload.
    - btw: kvass is running in production by public cloud provider, if all you need is prometheus, you should check it.
- jbd@ for mentioning this problem on twitter and reviewing early design doc with alolitas@.
- David Ashpole for suggesting using CRD sub resource for scaling and a very detailed write up in the comparing sts and
  ds issue.

## References

- [Design a Prometheus-specific CRD for the Operator](https://github.com/open-telemetry/wg-prometheus/issues/23) issue
  created by @jbd
- [Min's google doc for otel community](https://docs.google.com/document/d/13Gcu5SlbgjrsQJQUuZAjdQo1MOQA76Yji3oX8yHh-p8/edit#heading=h.l07w4n89sdn5)
  It also suggests how to extend exiting otel operator
