# driver-yeelight
Ninja Sphere driver (Go) for controlling Yeelight Sunflower bulbs 
(http://www.yeelight.com/en_US/product/yeelight-sunflower)

The driver finds a Yeelight hub using SSDP, identifies all lights and makes them available as things for the Ninja Sphere to control:

  - on/off state (brightness 100/0)
  - colour
  - brightness
  
Use the configuration (in Labs or http://ninjasphere.local) to:
 
  - control lights (on/off) directly
  - rename lights
  - create, delete and activate **presets/scenes** (collections of light states)
  - reset driver, clearing existing light bulbs
  - scan for and add new bulbs
  
If you have lights in the same room as your sphereamid then you will see two "pages" on the sphereamid - one for brightness and one for colour. Both of these can be adjusted using the airwheel gesture, and tapping on the brightness (first) page will toggle the light(s) on or off.

Installation
------------

Copy both package.json and the binary (from the release) into `/data/sphere/user-autostart/drivers/driver-yeelight` (create the directory as needed) and run `nservice driver-yeelight start` on (or restart) the sphereamid.

Known Issues
------------

There is no way yet in the Ninja Sphere system to "unexport" devices, so to remove a light, you will have to use the Yeelight hub, reset the lights through the config, then restart the driver or sphereamid. (This driver can update the config but not remove the devices without restarting.)

"Things" are not able to be renamed yet so adding lights to rooms in the phone app will show as IDs not your names. 
Workarounds: Rename things in the phone app or - rename your lights in Labs/ninjasphere.local, then stop the driver and delete the things using https://github.com/lindsaymarkward/sphere-thing-deleter (use the command `sphere-thing-deleter name Yee`) then re-run the driver. It will find the things again and use the names you set in the config.
