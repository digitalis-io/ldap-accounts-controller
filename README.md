# Kubernetes Controller for OpenLDAP accounts

> :warning: **This is just a learning exercise, not for real use**

## What is this?

This is just a learning exercise for me on using Operator. This code assumes you have an OpenLDAP server running in Kubernetes using the standard `posixAccount`. You can then use a simple yaml to create and delete users.

```yaml
apiVersion: ldap.digitalis.io/v1
kind: LdapUser
metadata:
  name: user01
spec:
  username: user01
  password: letmein
  gid: "1000"
  uid: "1000"
  homedir: /home/user01
  shell: /bin/bash
```

## Running

You can run it from command line using something like:

```sh
LDAP_BASE_DN="dc=axonops,dc=com" \
LDAP_BIND="cn=admin,dc=axonops,dc=com" \
LDAP_PASSWORD=xxxx \
LDAP_HOSTNAME=192.168.1.xx \
LDAP_PORT=30947 \
make install run
```

Or check out [kubebuilder](https://github.com/kubernetes-sigs/kubebuilder) docs on creating your own docker image to use inside kubernetes:

https://book.kubebuilder.io/quick-start.html#run-it-on-the-cluster

## Docker build

To deploy into a kubernetes you'll need to first create a docker image and push it to a registry from where k8s can download it. You can use for this:

```sh
make docker-build docker-push IMG=<some-registry>/<project-name>:tag
```

and after this you can deploy it with

```sh
make deploy IMG=<some-registry>/<project-name>:tag
```
