### E2E tests

```bash
make setup
make setup-deps
make install

```

Check for status, ready to go if all running

`watch kubectl get pods`

Port forwarding inorder to expose port for testing:

`make port-forward`
