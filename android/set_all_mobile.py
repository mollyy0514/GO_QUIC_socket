# run if the mobile phone has been turned off
# or if we want to git pull the latest version
from adbutils import adb
import os
import subprocess
import argparse

serial_to_device = {
    "R5CR20FDXHK":"sm00",
    "R5CR30P9Z8Y":"sm01",
    "R5CRA1GCHFV":"sm02",
    "R5CRA1JYYQJ":"sm03",
    "R5CRA1EV0XH":"sm04",
    "R5CRA1GBLAZ":"sm05",
    "R5CRA1ESYWM":"sm06",
    "R5CRA1ET22M":"sm07",
    "R5CRA1D23QK":"sm08",
    "R5CRA2EGJ5X":"sm09",
    "R5CRA1ET5KB":"sm10",
    "R5CRA1D2MRJ":"sm11",
}

device_to_port = {
    "sm00": [5200, 5201],
    "sm01": [5202, 5203],
    "sm02": [5204, 5205],
    "sm03": [4206, 4207],
    "sm04": [4208, 4209],
    "sm05": [4210, 4211],
    "sm06": [4212, 4213],
    "sm07": [4214, 4215],
    "sm08": [4216, 4217],
    "sm09": [4218, 4219],
}

parser = argparse.ArgumentParser()
parser.add_argument("-t", "--time", type=int,
                    help="time in seconds to transmit for (default 1 hour = 3600 secs)", default=300)
parser.add_argument("-b", "--bitrate", type=str,
                    help="target bitrate in bits/sec (0 for unlimited)", default="1M")
parser.add_argument("-l", "--length", type=int,
                    help="length of buffer to read or write in bytes (packet size)", default=250)
args = parser.parse_args()

devices_info = []
for i, info in enumerate(adb.list()):
    try:
        if info.state == "device":
            # <serial> <device|offline> <device name>
            devices_info.append((info.serial, info.state, serial_to_device[info.serial]))
        else:
            print("Unauthorized device {}: {} {}".format(serial_to_device[info.serial], info.serial, info.state))
    except:
        print("Unknown device: {} {}".format(info.serial, info.state))

devices_info = sorted(devices_info, key=lambda v:v[2])

devices = []
for i, info in enumerate(devices_info):
    devices.append(adb.device(info[0]))
    print("{} - {} {} {}".format(i+1, info[0], info[1], info[2]))
print("-----------------------------------")

tools = ["git", "iperf3m", "iperf3", "tcpdump", "tmux", "vim"]
for device, info in zip(devices, devices_info):
    # 抓最新版本的 wmnl-handoff-research
    print(info[2], device.shell("su -c 'cd /sdcard/wmnl-handoff-research && /data/git pull'"))
    print("-----------------------------------")
    if info[2][:2] == "sm":
        device.shell("su -c 'mount -o remount,rw /system/bin'")
        for tool in tools:
            device.shell("su -c 'cp /sdcard/wmnl-handoff-research/experimental-tools/android/sm-script/termux-tools/{} /bin'".format(tool))
            device.shell("su -c 'chmod +x /bin/{}'".format(tool))
        device.shell("su -c 'cp /data/python3 /bin'")
        device.shell("su -c 'chmod +x /bin/python3'")
        # GO environment setting
        # git pull the latest version and go build
        print(info[2], device.shell("su -c 'cd /data/data/com.termux/files/home/GO_QUIC_socket && /data/git restore . && /data/git pull'"))


    elif info[2][2] == "xm":
        # device.shell("su -c 'mount -o remount,rw /system/sbin'")
        for tool in tools:
            device.shell("su -c 'cp /sdcard/wmnl-handoff-research/experimental-tools/android/xm-script/termux-tools/{} /sbin'".format(tool))
            device.shell("su -c 'chmod +x /sbin/{}'".format(tool))
        device.shell("su -c 'cp /data/python3 /sbin'")
        device.shell("su -c 'chmod +x /sbin/python3'")
        # TODO: GO environment setting
    

# for info in devices_info:
#     print(info[0], info[2], "\n")
#     portString = f"{device_to_port[info[2]][0]},{device_to_port[info[2]][1]}"
#     su_cmd = f'cd /data/data/com.termux/files/home/GO_QUIC_socket && ./client_phone/client_socket.sh {info[2]} {portString} {args.time} {args.bitrate} {args.length}'
#     adb_cmd = f"su -c '{su_cmd}'"
#     p = subprocess.Popen([f'adb -s {info[0]} shell "{adb_cmd}"'], shell=True, preexec_fn = os.setpgrp)
#     # procs.append(p)

print('---End Of File---')
