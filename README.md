# 🚀 FUSE Challenge – NFS-Backed Read-Through Filesystem

[![Go](https://img.shields.io/badge/Go-1.20+-blue)](https://golang.org)  
[![FUSE](https://img.shields.io/badge/FUSE-v3-green)](https://github.com/libfuse/libfuse)  
[![Cache](https://img.shields.io/badge/Cache-LRU--Hashed-lightgrey)](https://en.wikipedia.org/wiki/Cache_replacement_policies)

A **read-only FUSE-based virtual filesystem** that overlays a simulated `nfs/` backend with a hashed, LRU-based `ssd/` cache, mounted at `./mnt/all-projects`.

---


## ✅ Features

📁 Mounts at ./mnt/all-projects

🪞 Mirrors nfs/ directory structure

⚡ Caching with content hash (SHA-256) for deduplication

🧠 LRU cache eviction: Max 10 files or 100KB

📥 Cache miss: read with 500ms delay and copy to ssd/

📤 Cache hit: instant read from ssd/

🚫 Read-only access (no writes allowed)

👋 Clean shutdown with umount, via SIGINT (Ctrl+C)

---

## 🛠️ Setup & Usage

```bash
# Create the nfs and ssd simulating dirs
python3 init_fs_environment.py

# Run
go run .

# In another terminal, see the results:
cat ./mnt/all-projects/project-1/common-lib.py   # expect 500ms delay
cat ./mnt/all-projects/project-1/common-lib.py   # instant (cache hit)
cat ./mnt/all-projects/project-2/common-lib.py   # instant (cache hit) as files are identical
ls ./mnt/all-projects
tree ./mnt/

# Run a script through the mount
cd ./mnt/all-projects/project-1 && python3 main.py
cd ../../../

# Cache eviction by file count, LRU, ten files
for i in {1..12}; do echo "file $i" > nfs/project-1/file-$i.py; done
for i in {1..12}; do cat ./mnt/all-projects/project-1/file-$i.py > /dev/null; done
ls ssd | wc -l # count files
ls ssd/

# Cache eviction by size, LRU, 100KB
mkdir -p nfs/project-evict-test
for i in {1..12}; do base64 /dev/urandom | head -c 20480 > nfs/project-evict-test/file-$i.dat; done
for i in {1..12}; do cat ./mnt/all-projects/project-evict-test/file-$i.dat > /dev/null; done
ls -lh ssd/ # Only the most recent files should remain in cache
du -sh ssd/

```