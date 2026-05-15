set -x
SIZE=29
FAVICO_SIZE=48
scale() {
    echo "$1/$SIZE" | bc -l
}
for size in 48 32 16 192 180 512; do
SCALE=$(scale "$size")
    for goje in goje goje-sad; do
        aseprite -b $goje.aseprite --scale "$SCALE" --save-as ../public/assets/$goje-${size}x${size}.png
    done
done
favico_scale=$(scale "$FAVICO_SIZE")

aseprite -b goje.aseprite --scale "$favico_scale"  --save-as ../public/favicon.ico
