## How to clone a private repository?
```
git clone https://<pat>@github.com/mollyy0514/GO_QUIC_socket.git
```
[Clone A Private Repository (Github)](https://stackoverflow.com/questions/2505096/clone-a-private-repository-github)

## How to import third party package?
1. Just import the package in your code
2. run `$ go mod init <filename>`, it will then create a go.mod file
3. run `$ go mod tidy`, this command checks the imports used in your program and fetches the module if not fetched already.


## TODO
**1031**:
- Datagram packet or Stream packet -> Stream packet
- Send packets under TCP protocol for 5 minutes

**1101**:
- Add time_sync
- Calculate latency

**1102**:
- So it turns out that i just need to calculate RTT, since TCP sends `ACK` for every packet, and in *The First 5G-LTE Comparative Study in Extreme Mobility*, It also compared the RTT between two protocols.
    - TCP: use `ACK Epoch time` minus the time that is send from client to server
    - QUIC: use `ACK Epoch time` minus the time that is send from client to server

**1103**:
- How to calculate QUIC RTT?
    - Add tcpdump in client
    - Read https://web.cs.ucla.edu/~lixia/papers/UnderstandQUIC.pdf
    - Does QUIC need to receive ACK and then it can send the next packet?
- Calculate TCP RTT

**11109**:
- Establish a new connection with server & client to make sure that both computers have got the latest tls key file
- TCP RTT:
    - https://www.freekb.net/Article?id=945
    - https://www.youtube.com/watch?v=Y5y85Lc7vOk&ab_channel=LauraChappell
- QUIC RTT:
    - (Rcv time - Send time) vs. Time since the ACKed packets
    - Adjust by ack_delay vs. No need to adjust since it's already the time delta

**1114**:
- Add tcpdump instructions in TCP & QUIC socket code.
- Design and observe packet loss situation.
- How to make throughput comparison?
- How to do congestion window(cwnd) camparison?

## Descrption

Comparison:
- Stream Socket:
    - Dedicated & end-to-end channel between server and client.
    - Use TCP protocol for data transmission.
    - Reliable and Lossless.
    - Data sent/received in the similar order.
    - Long time for recovering lost/mistaken data

- Datagram Socket:
    - Not dedicated & end-to-end channel between server and client.
    - Use UDP for data transmission.
    - Not 100% reliable and may lose data.
    - Data sent/received order might not be the same.
    - Don't care or rapid recovering lost/mistaken data.



## References
[使用Wireshark抓取quic-go產生的QUIC封包](https://hackmd.io/@pjkXMg3PTpu1Habe8pG_Lg/H103K8smn?utm_source=preview-mode&utm_medium=rec)
