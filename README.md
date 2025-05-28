# 🚀 FUSE Challenge – NFS-Backed Read-Through Filesystem

This project implements a **read-only FUSE-based virtual filesystem** that overlays a simulated `nfs/` storage backend with a readthrough `ssd/` cache.

Files are transparently fetched from the NFS directory with a simulated delay and cached to a shared SSD directory to improve subsequent access performance.

---

## ✅ Features

- Read-only FUSE mount at `./mnt`
- Directory listing reflects structure of `nfs/`
- Files are read through a cache:
  - If present in `ssd/`, read instantly
  - If not, read from `nfs/` **with a 500ms delay** and then cached
- Shared cache across all projects
- Clean unmount on exit

---

## 🛠️ Setup & Usage

### 1. Initialize the environment

```bash
# Create the nfs and ssd simulating dirs
python3 init_fs_environment.py

# Run the beautiful, beautiful code
go run main.go

# In another terminal, see the results:
ls ./mnt/all-projects
cat ./mnt/all-projects/project-1/common-lib.py   # expect 500ms delay
cat ./mnt/all-projects/project-2/common-lib.py   # no 500ms delay

