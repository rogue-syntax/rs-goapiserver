#!/bin/bash

# Check if exactly one argument is provided
if [ "$#" -ne 1 ]; then
  echo "Usage: $0 <major|minor>"
  exit 1
fi

# Get the current version tag
current_version=$(git describe --tags --abbrev=0)

# Parse the version tag into major, minor, and patch components
IFS='.' read -r -a version_parts <<< "${current_version//v/}"

major=${version_parts[0]}
minor=${version_parts[1]}
patch=${version_parts[2]}

# Increment the appropriate component
if [ "$1" == "major" ]; then
  minor=$((minor + 1))
  patch=0
elif [ "$1" == "minor" ]; then
  patch=$((patch + 1))
else
  echo "Invalid argument: $1. Use 'major' or 'minor'."
  exit 1
fi

# Construct the new version tag
new_version="v$major.$minor.$patch"

# Create the new version tag
git tag -a "$new_version" -m "Updated to version $new_version"

# Write the new version tag to version.txt
echo "$new_version" > version.txt

echo "Updated version to $new_version"
