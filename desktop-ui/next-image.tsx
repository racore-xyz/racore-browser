import type { ImgHTMLAttributes } from "react";

type StaticImageSource = { src: string };
type DesktopImageProps = Omit<ImgHTMLAttributes<HTMLImageElement>, "src"> & {
  src: string | StaticImageSource;
  priority?: boolean;
};

export default function Image({
  src,
  priority = false,
  loading,
  alt,
  ...props
}: DesktopImageProps) {
  const resolvedSource = typeof src === "string" ? src : src.src;

  return (
    <img
      {...props}
      src={resolvedSource}
      alt={alt ?? ""}
      loading={priority ? "eager" : (loading ?? "lazy")}
    />
  );
}
