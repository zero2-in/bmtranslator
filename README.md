# BMTranslator

Converts your BMS levels to zip archives that you can import into other rhythm games.

## How to Use

Output to Quaver with default volume of 100%: `bmt.exe -i /path/to/input -o /path/to/output`

Output to osu! with volume of 90% and more logs enabled: `bmt.exe -i /path/to/input -o /path/to/output -type osu -vol 90 --v`

**NOTE: Don't set the output (`-o`) to your game's Songs folder!** Create a separate output folder first, then import by dragging them into the game.

## Options

| Option | Description  | Default |
| ------------ | ------------ | ---- |
|  `-i` | Path of input folder containing FOLDERS of bms charts (NOT zip files!!!). If the folder is nested in another folder, just leave it, BMT will automatically handle that.  | N/A |
|  `-o` | Path to output the converted files to. | N/A |
|  `-vol` | Volume of hit sounds. (0-100) | 100 |
|  `-type` | Which type of file to convert to. You can choose `quaver` or `osu`. | quaver |
|  `-hp` | **osu! only.** Specify the HP drain rate. (0.0-10.0) | 8.5 |
|  `-od` | **osu! only.** Specify the overall difficulty. (0.0-10.0) | 8.0 |
|  `--v` | If this is specified, all logs (including some debug information) will be shown. Useful if you want to know why some maps didn't convert. | N/A |
|  `--keep-subtitles` | If this is specified, [implicit subtitles](https://hitkey.nekokan.dyndns.info/cmds.htm#TITLE-IMPLICIT-SUBTITLE) will NOT be removed from song titles. | N/A |
|  `--no-storyboard` | **osu! only.** If this is specified, background animation frames won't be parsed or inserted into the output files. | N/A |
|  `--no-measure-lines` | If this is specified, timing points will **not** be added at the end of each track to create visible measure lines. (It's a cosmetic thing and doesn't affect gameplay, but it might make slowjam unreadable. Some BMS files will appear unsnapped with timing lines if this is enabled.) | N/A |
|  `--dump-file-data` | If this is specified, raw file data will be dumped to a `.txt` file, which is put into the output folder. Each file will contain everything that BMTranslator knew about a BMS file. (Don't enable this unless you know what you're doing) | N/A |

## Credits

This tool wouldn't be possible without help from others. Thank you to:

- [Swan](https://github.com/Swan) for helping me with .qua files
- [yahweh](https://osu.ppy.sh/users/10465260) for moral support and for being a cool friend
- [mat](https://osu.ppy.sh/users/6283029) for helping me with the math side of things
- [Jakads](https://osu.ppy.sh/users/259972) for helping with 0 BPM change in osu
- [hitkey](https://hitkey.nekokan.dyndns.info) for their very helpful [BMS notes](https://hitkey.nekokan.dyndns.info/cmds.htm)
- Various people on the osu! forums who explained how timing points in osu! work.

## Limitations

- `#IF 1` will always be used. All other `#IF` blocks will be ignored.
- Almost no BMS maps use long notes in channels `51-59`, they use `#LNOBJ`. As a result, LNs placed in channels `51-59` are **untested**, but they are implemented. If you find a problem with them, please open an issue.
- BMS maps that use images as frames for the Background Animation can't be reliably parsed if the frames are <1ms apart, since osu! requires truncation of the decimal.
- If a BPM change occurs at any point within a STOP command, BMTranslator will still be able to parse the map, but the timing of the rest of the song will most likely be fucked. *However*, this has not appeared in a single map that I've tested, and by this reasoning, I think the only way to do this is by editing a BMS file by hand.