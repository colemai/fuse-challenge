# init_fs_environment.py

import os

# Define directory structure
base_dirs = {
    "ssd": [],
    "nfs/project-1": ["main.py", "common-lib.py"],
    "nfs/project-2": ["entrypoint.py", "common-lib.py"],
}

# Sample content for the test files
sample_content = {
    "main.py": 'print("Hello from main.py in project-1")\n',
    "entrypoint.py": 'print("Starting project-2 entrypoint")\n',
    "common-lib.py": 'def util(): return "Shared util function"\n',
}

# Create directories and files
for dir_path, files in base_dirs.items():
    os.makedirs(dir_path, exist_ok=True)
    for file_name in files:
        full_path = os.path.join(dir_path, file_name)
        with open(full_path, "w") as f:
            f.write(sample_content[file_name])

# Create mount point
os.makedirs("./mnt/all-projects", exist_ok=True)

print(" /mnt ready for FUSE mount")
print("✅ Directory structure and test files created.")
