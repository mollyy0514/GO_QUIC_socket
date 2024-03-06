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

**1117**:
- Modulize socket code, so that it can automatically organize the experiment result. Especially tls_key.log for QUIC need to be stored in order to check previous experiment data in Wireshark.

**1120**:
- ==Implement qlog & qvis==

**1122**:
- Connect to devices and assign ports

**1127**:
- Open sockets for every UDP phone

**1204**:
- Open sockets for every UDP phone
- Fix pop out EOF problem
- Run concurrent program
- So the procedure is: run concurrently on multiple ports -> check adb works well 

**1213**:
- Update phone serials
- Test 2 phones

**1228**:
- 0-rtt implementation

**0111**:
- Add dl
- Mobile phone experiment program

**0118**:
- Analyze experiment data
    - MobileInsight file
    - .qlog to .json
- Will out-of-order occur in QUIC?
    - Since QUIC is build on top of UDP, out-of-order won't lead to HOL blocking.
- What's bytes-in-flight?
    - The bytes that have been sent but not yet ACKed, if the cwnd = 64k & bytes-in-flight = 48k, then it can only be sent 16k more before the rwnd is filled.

**0305**:
- Modify the bitrate calculation to float64.
- Create a new python file to trigger client socket.
- Modify the time file to .txt.

**0306**:
- Write a file upload socket.