# Eventrouter

Extremely simplified fork of <https://github.com/heptiolabs/eventrouter>

Contains only the stdout sink

## Running Eventrouter

Startup:

```bash
kubectl create -f https://raw.githubusercontent.com/mwennrich/eventrouter/main/yaml/eventrouter.yaml
```

Teardown:

```bash
kubectl delete -f https://raw.githubusercontent.com/mwennrich/eventrouter/main/yaml/eventrouter.yaml
```

### Inspecting the output

```bash
kubectl logs -f deployment/eventrouter -n kube-system
```
