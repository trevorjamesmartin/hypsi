# hypsi
a simple [hyprpaper](https://wiki.hyprland.org/Hypr-Ecosystem/hyprpaper/) management tool, written in Go

now with optional webview interface

This program started as a script to manage my desktop wallpaper between sway and hyprland. When I first started using [Hyprland](https://www.hyprland.org/), I was also jumping in and out of sway and using [swww](https://github.com/LGFae/swww) to set the wallpaper on both.

When I found hypaper being developed by the author of Hyprland I had to give it a try... It lives up to it's claim of being 'blazing fast' perhaps in part by not overdoing it.
- You can read this [important note to the inner workings (of hyprpaper)](https://github.com/hyprwm/hyprpaper#important-note-to-the-inner-workings) to unpack that.
- TLDR; hyprpaper gives you full control of wallpaper management

Assuming you have Hyprland installed and you've enabled the hyprpaper plugin, you can
- `hypsi /path/to/imagefile.jpg` set your wallpaper from the command line and have it persist
- `hypsi %f` create a custom action in Thunar (enabling you to right-click an image > set as wallpaper)

If you have rofi istalled, try running
- `PREVIEW=true rofi -mode filebrowser -show filebrowser|xargs hypsi` from a folder containing images

![rofi integration](./rofi-integration.jpg)

Now this little Go app provides a few additional crossover features for web devs to play with
- `hypsi -listen` to start a local web server (change your desktop wallpaper over the network)
- `hypsi -json` to show your hyprpaper configuration in simple JSON format
- `hypsi -html` to render the HTML without starting a server
- `hypsi -webview` open the optional webview interface

![screenshot 3](./screenshot3.jpg)
![screenshot 2](./screenshot2.jpg)
![screenshot 1](./screenshot1.jpg)

# templating

In spirit with Hyprland, hypsi is highly configuable & you may override the default template.

To get started customizing your template, you'll need to copy the base template.
- running `hypsi -develop` will write the base template files to your current working directory, if they do not already exist.


[on my NixOS configuration](https://github.com/trevorjamesmartin/nixos-config) I install this [along with hyprpaper](https://github.com/trevorjamesmartin/nixos-config/tree/main/nixos/modules/home-manager/hyprpaper) using a module system. Your system integration may differ from mine, of course these are all just suggestions. You can even build the app in one command with `nix build github:trevorjamesmartin/hypsi`


feel free to fork this and modify for your system or project
