for size in 48x48 32x32 16x16 192x192 180x180; do
  magick goje-square.png -resize $size ../public/assets/goje-$size.png
done
