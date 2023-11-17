# <---------->
# 1. do git push and pull to get the server side data
# 2. make a folder with the date
# (if there are more than one experiment during that day, then store as {date}_i)
# 3. store timefloats_{date}.json & capturequic_c_{date}.pcap & cpaturequic_s_{date}.pcap into the created folder
# 4. move tls_key.log created by server side to this folder
# (while checking the file in Wireshark, press <cmd>+<,> to open Preferences -> Protocols -> TLS -> change (Pre)-Master-Secret log filename)
# 5. remember to git push and pull again
# <---------->
import re
import os
import sys
import shutil
from pathlib import Path
from datetime import date

def make_folder(date):
    Path(f"./data/{date}").mkdir(parents=True, exist_ok=True)

def move_file(prev, date, s, c, f, k):
    # s: serverside data, c: client side date, f: the time data in payload, k: tls_ley.log
    shutil.move(prev+s, prev+date+"/"+s)
    shutil.move(prev+c, prev+date+"/"+c)
    shutil.move(prev+f, prev+date+"/"+f)
    os.rename(prev+date+"/"+f, prev+date+f"/timefloats_{date}.json")
    shutil.move(k, prev+date+"/"+k)

date = sys.argv[1]
# make a folder to store the ecperiment data
make_folder(date)
date = date[:8]
s = f"capturequic_s_{date}.pcap"
c = f"capturequic_c_{date}.pcap"
f = "timefloats.json"
move_file("./data/", date, s, c, f, "tls_key.log")