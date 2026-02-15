import os
import re
import sys

def fix_imports(root_dir, package_prefix="lib.model"):
    print(f"--- Fixing Protobuf imports in: {root_dir} ---")

    # regex for versioned imports: from auth.v1 -> from lib.model.auth.v1
    # Excludes google, grpc, and the prefix itself to avoid double-fixing
    ver_regex = re.compile(rf'from (?!(google|grpc|{package_prefix.split(".")[0]}))([\w\.]+)\.v(\d+\w*)')

    # regex for tagger import: from tagger import -> from lib.model.third_party.tagger import
    tagger_regex = re.compile(r'^from tagger import', re.MULTILINE)

    # regex for same-directory imports: import auth_pb2 -> from . import auth_pb2
    rel_regex = re.compile(r'^import (.*_pb2)', re.MULTILINE)

    # regex for gRPC alias imports
    grpc_regex = re.compile(r'^import (.*_pb2) as', re.MULTILINE)

    count = 0
    for root, _, files in os.walk(root_dir):
        for f in files:
            if f.endswith(".py"):
                path = os.path.join(root, f)
                with open(path, 'r', encoding='utf-8') as f_in:
                    content = f_in.read()

                # Apply transformations
                new_content = ver_regex.sub(rf'from {package_prefix}.\2.v\3', content)
                new_content = tagger_regex.sub(rf'from {package_prefix}.tagger import', new_content)
                new_content = rel_regex.sub(r'from . import \1', new_content)
                new_content = grpc_regex.sub(r'from . import \1 as', new_content)

                if new_content != content:
                    with open(path, 'w', encoding='utf-8') as f_out:
                        f_out.write(new_content)
                    count += 1
    
    print(f"Done - Processed {count} files.")

if __name__ == "__main__":
    # Allow passing the directory as an argument, default to current dir
    target_dir = sys.argv[1] if len(sys.argv) > 1 else "."
    fix_imports(target_dir)