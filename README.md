# Image Dups

imgdups groups images from a given folder based on their [perceptual hash](https://en.wikipedia.org/wiki/Perceptual_hashing). 

## Usage
```bash
$ imgdups -h
Usage of imgdups:
  -dir string
        Images folder path
  -quiet
        If true, wont print the removed duplicates (default false)
  -workers int
        Number of workers to run concurrently (default 100)
```

For a given folder structure containing images, such as:
```bash
$ tree ss
ss
├── img1.png
├── img2.png
├── img3.png
├── img4.png
[...]

0 directories, 87 files
```

After running `imgdups` the images will be grouped by its perceptual hash into subfolders, leaving the first image associated with each hash in the root folder for easy review.

```bash
$ imgdups -dir ss/ -workers 50 -quiet
$ tree ss
ss
├── dups
│   ├── 0000000000000000
│   │   ├── img5.png
│   │   ├── img7.png
│   │   └── img11.png
│   ├── 0c08000000000000
│   │   └── img6.png
│   ├── c000000000000000
│   │   ├── img8.png
│   │   ├── img10.png
│   └── f177700000000000
│   │   ├── img9.png
[...]
├── img1.png
├── img2.png
├── img3.png
├── img4.png

7 directories, 87 files
```
