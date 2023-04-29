## Mino Encryption Channel

### Usage

```bash
mino -conf mino.yml
```

### Feature

- support transport
  - [x] tcp
  - [x] quic
- encoding
  - [x] xxor

### Config

```yaml
log_file: "mino.log"
channel:
  - input: tcp://:8080
    output: tcp://:8081?enc=xxor&key=mino
  - input: tcp://:8081?enc=xxor&key=mino
    output: tcp://:80
```