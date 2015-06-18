# sphere-yeelight
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
  
Installation
------------

Copy both package.json and the sphere-yeelight binary (from the release) into `/data/sphere/user-autostart/drivers/sphere-yeelight` (create the directory as needed) and run `nservice sphere-yeelight start` on (or restart) the sphereamid.

Known Issues
------------

There is no way yet in the Ninja Sphere system to "unexport" devices, so to remove a light, you will have to use the Yeelight hub, reset the lights through the config, then restart the driver or sphereamid. (This driver can update the config but not remove the devices without restarting.)

"Things" are not able to be renamed yet so adding lights to rooms in the phone app will show as IDs not your names. 
Workaround: Rename your lights, then stop the driver and delete the things using https://github.com/lindsaymarkward/sphere-thing-deleter then re-run the driver. It will find the things again and use the names you set in the config.
