#!/usr/bin/sh
last_tag=$(git tag | tail -n 1)
git tag "$1" || {
  last_tag=$(git tag | tail -n 2 | head -n 1)
  git checkout "$1"
}
files="goje_linux_amd64 goje_windows_amd64.exe goje_android_arm64"
make coverage.out $files
zip_files=""
for file in $files; do
  case "$file" in 
    *.exe)
      out="${file%.*}.zip"
      tmp=goje.exe
      mv "$file" $tmp
      zip "$out" $tmp "./goje-launcher.bat" || exit 1;;
    *)
    out="$file.bz2"
    tmp=goje
    mv "$file" $tmp
    bzip2 -c $tmp > "$out" || exit 1
  esac
  mv "$tmp" "$file"
  zip_files="$zip_files $out"
done
gh release create "$1" $zip_files  --title "$1" --notes "**Full Changelog**: https://github.com/nimaaskarian/goje/compare/$last_tag...$1" --repo nimaaskarian/goje
rm ${files} ${zip_files}
git checkout master
