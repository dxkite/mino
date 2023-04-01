## Mino Encryption Channel

### Usage

```bash
mino -conf mino.yml
```


### Config

```yaml
log_file: "mino.log"
tcp_channel:
  - input: tcp://:8080
    output: tcp://:1080?enc=xxor&key=mino
  - input: tcp://:1080?enc=xxor&key=mino
    output: tcp://:80
```