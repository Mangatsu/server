# 📂 Directory structure
**Multiple root directories are supported.** I suggest creating a structured format for proper long-running manga, and a freeform structure for doujinshi and other art collections. Examples follow:

- **Freeform**: galleries can be up to three levels deep. Good for doujinshi, one-shots and other more unstructured collections.
    - External JSON metadata files have to be placed in the same level as the gallery archive. Preferably having the same name as the gallery archive. If no exact match is found, filename close enough will be used instead.

```
📂 freeform
├── 📂 doujinshi
│       ├──── 📂 oppai
│       │     ├──── 📦 [Group (Artist)] Ecchi Doujinshi.cbz
│       │     └──── 📄 [Group (Artist)] Ecchi Doujinshi.json
│       ├──── 📦 (C99) [Group (Artist)] elfs.zip
│       ├──── 📄 (C99) [Group (Artist)] elfs.json
│       └──── 📦 (C88) [kroup (author, another author)] Tankoubon [DL].zip  (JSON or TXT metafile inside)
├── 📂 art
│       ├──── 📂 [Artist] Pixiv collection
│       │     ├──── 🖼️ 0001.jpg
│       │     ├────...
│       │     └──── 🖼️ 0300.jpg
│       ├──── 📦 art collection y.rar
│       └──── 📄 art collection y.json
└── 📦 (C93) [group (artist)] Lonely doujinshi (Magical Girls).cbz
```

- **Structured**: galleries follow a strict structure. Good for long-running manga (shounen, seinen etc).
    - `Manga -> Volumes -> Chapters`, `Manga -> Volumes` or `Manga -> Chapters`
    - 'Series' will be set to the name of the 1st level directory except for galleries in the root directory.

```
📂 structured
├── 📕 Manga 1
│       ├── 📦 Volume 1.cbz
│       ├── 📦 Volume 2.cbz
│       ├── 📦 Volume 3.cbz
│       └── 📦 Volume 4.zip
├── 📘 Manga 2
│       └── 📂 Vol. 1
│           ├── 🖼️ 0001.jpg
│           ├── ...
│           └── 🖼️ 0140.jpg
├── 📘 Manga 3
│       └── 📂 Vol. 1
│           ├── 📦 Chapter 1.zip
│           ├── 📦 Chapter 2.zip
│           └── 📦 Chapter 3.rar
├── 📗 Manga 4
│       ├── 📦 Chapter 1.zip
│       ├── ...
│       └── 📦 Chapter 30.rar
└── 📦 One Shot Manga.rar
```
