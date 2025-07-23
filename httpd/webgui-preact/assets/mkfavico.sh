for size in 48x48 32x32 16x16 192x192 180x180 512x512; do
  magick goje-square.png -resize $size ../public/assets/goje-$size.png
done
magick goje-square.png -resize 48x48 ../public/favicon.ico
