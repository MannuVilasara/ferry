## Ferry - Lightweight go written load balancer and reverse proxy

This is just a learning project to understand how load balancers actually distribute traffic.

### Tasks

- [x] Route Traffic to Single url
- [x] Implement Round Robin Algorithm
- [x] Add Health checks
- [x] Hot Reloading
- [x] Config Validation (Prevent crashes on bad configs)
- [x] CLI Tool Architecture (`ferry -r` to reload, `ferry -c` to check)
- [x] Admin API & CLI List (`ferry -l` to view active backends in a table)
- [ ] Systemd Service (Background daemonization)
- [ ] Implement Least Connections Algorithm
- [ ] Implement Weighted Round Robin
- [ ] Add Metrics
