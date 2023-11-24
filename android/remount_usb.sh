sudo rmmod pl2303
sudo rmmod option
sudo rmmod usb_wwan
sudo rmmod usbserial
sleep 1

sudo modprobe usbserial vendor=0x05c6 product=0x9091
sleep 1

adb devices
ls /dev/serial/by-id