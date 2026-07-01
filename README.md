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
- [x] Systemd Service (Background daemonization)
- [x] Implement Least Connections Algorithm
- [x] Implement Weighted Round Robin
- [x] Add Metrics
- [ ] Implement IP Hashing / Sticky Sessions
- [ ] Add Active Circuit Breaking & Retries
- [ ] Implement Rate Limiting (Token Bucket)
- [ ] Support TLS Termination (HTTPS)
