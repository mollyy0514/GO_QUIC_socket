#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import os
import time
import signal
import argparse
import subprocess
import datetime as dt

parser = argparse.ArgumentParser()
parser.add_argument("-d", "--device", type=str, help="device", default="sm00")
parser.add_argument("-p", "--ports", type=str, help="ports", default="5200,5201")
args = parser.parse_args()

dev = args.device
ports = args.ports.split(',')

now = dt.datetime.today()
n = [str(x) for x in [now.year, now.month, now.day, now.minute, now.second]]
n = [x.zfill(2) for x in n]
n = ''.join(n[:3]) + '_' + ''.join(n[3:])
pcap_path = "/sdcard/experiment_log/"
if not os.path.isdir(pcap_path):
    os.system(f'mkdir {pcap_path}')

pcapUl = os.path.join(pcap_path, f"client_pcap_UL_{dev}_{ports[0]}_{n}.pcap")
tcpprocUl = subprocess.Popen([f"tcpdump -i any port {ports[0]} -w {pcapUl}"], shell=True, preexec_fn=os.setpgrp)

pcapDl = os.path.join(pcap_path, f"client_pcap_DL_{dev}_{ports[1]}_{n}.pcap")
tcpprocDl = subprocess.Popen([f"tcpdump -i any port {ports[1]} -w {pcapDl}"], shell=True, preexec_fn=os.setpgrp)

stop_threads = False
while not stop_threads:
    try:
        time.sleep(3)

    except KeyboardInterrupt:
        print("Manually interrupted.")
        stop_threads = True

os.killpg(os.getpgid(tcpprocUl.pid), signal.SIGTERM)
os.killpg(os.getpgid(tcpprocDl.pid), signal.SIGTERM)
print(f"---Kill {dev} tcpdump.---")