param(
  [string]$ProjectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

Add-Type -AssemblyName System.Drawing

$outputDir = Join-Path $ProjectRoot "src-tauri\windows\nsis"
$logoPath = Join-Path $ProjectRoot "src-tauri\windows\nsis\racore-logo-transparent.png"
[System.IO.Directory]::CreateDirectory($outputDir) | Out-Null

function New-Canvas([int]$width, [int]$height) {
  return [System.Drawing.Bitmap]::new(
    $width,
    $height,
    [System.Drawing.Imaging.PixelFormat]::Format24bppRgb
  )
}

function New-Brush([string]$hex) {
  return [System.Drawing.SolidBrush]::new(
    [System.Drawing.ColorTranslator]::FromHtml($hex)
  )
}

$logo = [System.Drawing.Image]::FromFile($logoPath)
$paper = [System.Drawing.ColorTranslator]::FromHtml("#F7FAF5")
$paleGreen = [System.Drawing.ColorTranslator]::FromHtml("#EAF7D8")
$lime = [System.Drawing.ColorTranslator]::FromHtml("#82E600")
$dark = [System.Drawing.ColorTranslator]::FromHtml("#0C2725")

$header = New-Canvas 150 57
$headerGraphics = [System.Drawing.Graphics]::FromImage($header)
$headerGraphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$headerGraphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
$headerGradient = [System.Drawing.Drawing2D.LinearGradientBrush]::new(
  [System.Drawing.Rectangle]::new(0, 0, 150, 57),
  $paper,
  $paleGreen,
  0
)
$headerGraphics.FillRectangle($headerGradient, 0, 0, 150, 57)
$headerGraphics.DrawImage($logo, 7, 7, 136, 37)
$headerCaption = [System.Drawing.Font]::new("Segoe UI", 4.6, [System.Drawing.FontStyle]::Bold)
$headerGraphics.DrawString("LOCAL AI / PRIVATE BROWSING", $headerCaption, (New-Brush "#356B46"), 36, 43)
$headerGraphics.DrawLine([System.Drawing.Pen]::new($lime, 2), 7, 52, 143, 52)
$header.Save(
  (Join-Path $outputDir "header.bmp"),
  [System.Drawing.Imaging.ImageFormat]::Bmp
)

$sidebar = New-Canvas 164 314
$sidebarGraphics = [System.Drawing.Graphics]::FromImage($sidebar)
$sidebarGraphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$sidebarGraphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
$sidebarGradient = [System.Drawing.Drawing2D.LinearGradientBrush]::new(
  [System.Drawing.Rectangle]::new(0, 0, 164, 314),
  $paper,
  $paleGreen,
  90
)
$sidebarGraphics.FillRectangle($sidebarGradient, 0, 0, 164, 314)
$sidebarGraphics.FillEllipse((New-Brush "#D9F4B9"), -55, 178, 180, 180)
$sidebarGraphics.DrawEllipse([System.Drawing.Pen]::new($lime, 1), -35, 198, 140, 140)
$sidebarGraphics.DrawImage($logo, 12, 24, 140, 45)

$smallFont = [System.Drawing.Font]::new("Segoe UI", 6.5, [System.Drawing.FontStyle]::Regular)
$labelFont = [System.Drawing.Font]::new("Consolas", 7, [System.Drawing.FontStyle]::Bold)
$sidebarGraphics.DrawString("THE LOCAL-FIRST BROWSER", $smallFont, (New-Brush "#356B46"), 23, 92)
$sidebarGraphics.DrawString("PRIVATE / LOCAL / AGENTIC", $smallFont, (New-Brush "#5C8F24"), 23, 111)
$sidebarGraphics.DrawLine([System.Drawing.Pen]::new($lime, 2), 23, 137, 140, 137)

$features = @(
  @("01", "HAMMER 0.5B"),
  @("02", "LLAMA.CPP / CPU"),
  @("03", "RACORE MESH")
)
$featureY = 158
foreach ($feature in $features) {
  $sidebarGraphics.DrawString($feature[0], $smallFont, (New-Brush "#5C8F24"), 23, $featureY)
  $sidebarGraphics.DrawString($feature[1], $labelFont, (New-Brush "#0C2725"), 48, ($featureY - 1))
  $featureY += 27
}

$sidebarGraphics.DrawString("NO CLOUD MODEL KEYS", $smallFont, (New-Brush "#356B46"), 23, 272)
$sidebarGraphics.DrawString("RACORE.XYZ / 2026", $smallFont, (New-Brush "#5C8F24"), 23, 289)
$sidebar.Save(
  (Join-Path $outputDir "sidebar.bmp"),
  [System.Drawing.Imaging.ImageFormat]::Bmp
)

$headerGraphics.Dispose()
$headerGradient.Dispose()
$header.Dispose()
$sidebarGraphics.Dispose()
$sidebarGradient.Dispose()
$sidebar.Dispose()
$logo.Dispose()

Write-Output "Generated NSIS assets in $outputDir"
