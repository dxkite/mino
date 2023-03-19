## Mino Encryption Channel

### Usage

**Client**
```
mino -input tcp://:8080 -output tcp://:1080?enc=xxor&key=mino
```

**Server**
```
mino -input tcp://:1080?enc=xxor&key=mino -output  tcp://:80
```