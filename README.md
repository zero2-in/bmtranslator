# BMTranslator

Converts your BMS levels to zip archives that you can import into other rhythm games.

## How to Use

Output to quaver: `bmt.exe -i /path/to/input -o /path/to/output`

Output to osu! with volume of 90% and more logs enabled: `bmt.exe -i /path/to/input -o /path/to/output -type osu -vol 90 --v`

**NOTE: Don't set the output (`-o`) to your game's Songs folder!** Create a separate output folder first, then import by dragging them into the game.

## Options

| Option | Description  | Default |
| ------------ | ------------ | ---- |
|  `-i` | Path of input folder containing FOLDERS of bms charts (NOT zip files!!!). If the folder is nested in another folder, just leave it, BMT will automatically handle that.  | N/A |
|  `-o` | Path to output the converted files to. | N/A |
|  `-vol` | Volume of hit sounds. Range is 0 to 100. | 100 |
|  `-type` | Which type of file to convert to. You can choose `quaver` or `osu`. | quaver |
|  `-hp` | **osu! only.** Specify the HP drain rate. | 8.5 |
|  `-od` | **osu! only.** Specify the overall difficulty. | 8.0 |
|  `--v` | If this is specified, all logs (including some debug information) will be shown. | N/A |
|  `--no-brackets` | If this is specified, ALL brackets `[]` and their contents will be removed from the song's title. | N/A |

## Credits

This tool wouldn't be possible without help from others. Thank you to:

- [Swan](https://github.com/Swan) for helping me with .qua files
- [yahweh](https://osu.ppy.sh/users/10465260) for moral support and for being a cool friend
- [mat](https://osu.ppy.sh/users/6283029) for helping me with the math side of things
- [Jakads](https://osu.ppy.sh/users/259972) for helping with 0 BPM change in osu
- Various people on the osu! forums who explained how timing points in osu! work.

## Limitations

Right now, BMT can *not* parse `#RANDOM` headers, which means if they are in a map, the converter *will use ALL track values, regardless of whether it exists in a `#IF n` or not!* (This is not permanent; it will be parsed eventually)

Almost no BMS maps use long notes in channels `51-59`, they use `#LNOBJ`. As a result, LNs placed in channels `51-59` are **untested**, but they are implemented. If you find a problem with them, please open an issue.

BMS maps that use images as frames for the Background Animation can't be reliably parsed if the frames are <1ms apart, since osu! requires truncation of the decimal.