#!/bin/bash
cd backend

if [ $# -eq 0 ]; then
  # No arguments, format all files
  go fmt ./...
else
  # Arguments provided, format specific files
  # Convert paths to be relative to backend directory
  relative_files=()
  for file in "$@"; do
    # If file starts with backend/, remove the backend/ prefix
    if [[ "$file" == backend/* ]]; then
      relative_file="${file#backend/}"
    # If file is an absolute path, extract the part after backend/
    elif [[ "$file" == /* ]] && [[ "$file" == */backend/* ]]; then
      relative_file="${file#*/backend/}"
    else
      # Assume it's already relative to backend
      relative_file="$file"
    fi
    relative_files+=("$relative_file")
  done

  # Group files by directory since go fmt requires all files in same directory
  declare -A dir_groups
  for file in "${relative_files[@]}"; do
    dir=$(dirname "$file")
    if [[ "$dir" == "." ]]; then
      dir=""
    fi
    if [[ -z "${dir_groups[$dir]}" ]]; then
      dir_groups[$dir]="$file"
    else
      dir_groups[$dir]="${dir_groups[$dir]} $file"
    fi
  done

  # Format each group of files
  for dir in "${!dir_groups[@]}"; do
    if [[ -z "$dir" ]]; then
      # Files in root directory
      go fmt ${dir_groups[$dir]}
    else
      # Files in subdirectory - format the directory
      go fmt "./$dir"
    fi
  done
fi
