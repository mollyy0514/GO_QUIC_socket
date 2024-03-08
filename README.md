## How to clone a private repository?
```
git clone https://<pat>@github.com/mollyy0514/GO_QUIC_socket.git
```
[Clone A Private Repository (Github)](https://stackoverflow.com/questions/2505096/clone-a-private-repository-github)


## How to import third party package?
1. Just import the package in your code
2. run `$ go mod init <directory>`, it will then create a go.mod file
3. run `$ go mod tidy`, this command checks the imports used in your program and fetches the module if not fetched already.


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
