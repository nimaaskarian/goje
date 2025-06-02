last_tag=$(git tag | tail -n 1)
git tag "$1" || {
  last_tag=$(git tag | tail -n 2 | head -n 1)
  git checkout "$1"
}
make
files=(goje_linux_amd64 goje_windows_amd64.exe goje_android_arm64)
zip_files=()
for file in "${files[@]}"; do
  if [[ $file = *.exe ]]; then
    out="${file%.*}.zip"
    zip "$out" "$file" "./goje-launcher.bat"
  else
    out="$file.bz2"
    bzip2 -c "$file" > "$out"
  fi
  zip_files+=("$out")
done
gh release create "$1" "${zip_files[@]}"  --title "$1" --notes "**Full Changelog**: https://github.com/nimaaskarian/goje/compare/$last_tag...$1" --repo nimaaskarian/goje
rm "${files[@]}" "${zip_files[@]}"
git checkout master
