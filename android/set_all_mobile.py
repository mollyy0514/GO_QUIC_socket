# run if the mobile phone has been turned off
# or if we want to git pull the latest version
from adbutils import adb
# "R5CRA1ET5KB":"sm00",
serial_to_device = {
    "R5CR20FDXHK":"sm00",
    "R5CRA1D2MRJ":"sm01",
    "R5CRA1GCHFV":"sm02",
    "R5CRA1JYYQJ":"sm03",
    "R5CRA1EV0XH":"sm04",
    "R5CRA1GBLAZ":"sm05",
    "R5CRA1ESYWM":"sm06",
    "R5CRA1ET22M":"sm07",
    "R5CRA1D23QK":"sm08",
    "R5CRA2EGJ5X":"sm09",
    "73e11a9f":"xm00",
    "491d5141":"xm01",
    "790fc81d":"xm02",
    "e2df293a":"xm03",
    "28636990":"xm04",
    "f8fe6582":"xm05",
    "d74749ee":"xm06",
    "10599c8d":"xm07",
    "57f67f91":"xm08",
    "232145e8":"xm09",
    "70e87dd6":"xm10",
    "df7aeaf8":"xm11",
    "e8c1eff5":"xm12",
    "ec32dc1e":"xm13",
    "2aad1ac6":"xm14",
    "64545f94":"xm15",
    "613a273a":"xm16",
    "fe3df56f":"xm17",
}

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
            device.shell("su -c 'cp /sdcard/wmnl-handoff-research/experimental-tools-beta/android/sm-script/termux-tools/{} /bin'".format(tool))
            device.shell("su -c 'chmod +x /bin/{}'".format(tool))
        device.shell("su -c 'cp /data/python3 /bin'")
        device.shell("su -c 'chmod +x /bin/python3'")
        # GO environment setting
        # git pull the latest version and go build
        print(info[2], device.shell("su -c 'cd /data/data/com.termux/files/home/GO_QUIC_socket && /data/git pull'"))
        device.shell("su -c 'cd /data/data/com.termux/files/home/GO_QUIC_socket && ./client_phone/client_socket.sh'")

    elif info[2][2] == "xm":
        # device.shell("su -c 'mount -o remount,rw /system/sbin'")
        for tool in tools:
            device.shell("su -c 'cp /sdcard/wmnl-handoff-research/experimental-tools-beta/android/xm-script/termux-tools/{} /sbin'".format(tool))
            device.shell("su -c 'chmod +x /sbin/{}'".format(tool))
        device.shell("su -c 'cp /data/python3 /sbin'")
        device.shell("su -c 'chmod +x /sbin/python3'")
        # TODO: GO environment setting
    
    # # test tools
    # print(info[2], 'iperf3m:', device.shell("su -c 'iperf3m --version'"))
    # print("-----------------------------------")
    
    # # UDP_Phone
    # # 覆寫掉原本的 UDP_Phone
    # su_cmd = 'rm -rf /sdcard/UDP_Phone && cp -r /sdcard/wmnl-handoff-research/experimental-tools-beta/udp-socket-programming/v3/UDP_Phone /sdcard'
    # adb_cmd = f"su -c '{su_cmd}'"
    # device.shell(su_cmd)
    # print(info[2], 'Update UDP_Phone! v3')
    # print("-----------------------------------")
    
    # # TCP_Phone
    # # 覆寫掉原本的 TCP_Phone
    # su_cmd = 'rm -rf /sdcard/TCP_Phone && cp -r /sdcard/wmnl-handoff-research/experimental-tools-beta/tcp-socket-programming/v3/TCP_Phone /sdcard'
    # adb_cmd = f"su -c '{su_cmd}'"
    # device.shell(su_cmd)
    # print(info[2], 'Update TCP_Phone! v3')
    # print("-----------------------------------")

print('---End Of File---')
