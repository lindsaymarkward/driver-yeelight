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
  - create, delete and activate presets/scenes (collections of light states)
  - reset driver, clearing existing light bulbs
  - scan for and add new bulbs
  