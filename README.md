## Mino Encryption Channel

### Usage

```bash
mino -conf mino.yml
```

### Feature

- [x] tcp transport
- [ ] quic transport in processing

### Config

```yaml
log_file: "mino.log"
channel:
  - input: tcp://:8080
    output: tcp://:8081?enc=xxor&key=mino
  - input: tcp://:8081?enc=xxor&key=mino
    output: tcp://:80
```