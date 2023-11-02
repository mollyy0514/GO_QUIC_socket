import ntplib
import time
import sys
import socket
import datetime as dt
from myutils import makedir

HOST = '192.168.1.78'
PORT = 3298

def clock_diff():
    
    return 

# client
if sys.argv[1] == '-c':
    now = dt.datetime.today()
    date = [str(x) for x in [now.year, now.month, now.day]]
    date = [x.zfill(2) for x in date]
    date = '-'.join(date)
    dirpath = f'./log/{date}'
    makedir(dirpath)

    server_addr = (HOST, PORT)
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    outdata = server_addr
    s.sendto(outdata.encode(), server_addr)
    s.settimeout(3)
    try:
        indata, addr = s.recvfrom(1024)
        time1 = time.time()
        indata = indata.decode()
    except:
        print("timeout", outdata)
    outdata = 'end'
    s.sendto(outdata.encode(), server_addr)


elif sys.argv[1] == '-s':
    s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
    s.bind(('0.0.0.0', PORT))

    while True:
        print('server start at: %s:%s' % ('0.0.0.0', PORT))
        print('wait for connection...')

        while True:
            indata, addr = s.recvfrom(1024)
            indata = indata.decode()
            if indata == 'end':
                print(indata)
                break
            print('recvfrom ' + str(addr) + ': ' + indata)
            
            server_time_now = time.time()
            outdata = f'{server_time_now}'
            s.sendto(outdata.encode(), addr)
            print('server send' + indata, server_time_now)


# Get the time difference
time_difference = clock_diff()

# Store the time difference in a file
with open('time_difference.txt', 'w') as file:
    file.write(f"The time difference between server and client is {time_difference} seconds.")

print("Time difference saved to 'time_difference.txt' file.")
