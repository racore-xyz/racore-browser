param(
  [string]$ProjectRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
)

Add-Type -AssemblyName System.Drawing

$outputDir = Join-Path $ProjectRoot "src-tauri\windows\nsis"
$iconPath = Join-Path $ProjectRoot "src-tauri\icons\icon.png"
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

$logo = [System.Drawing.Image]::FromFile($iconPath)
$navy = [System.Drawing.ColorTranslator]::FromHtml("#06111D")
$deepBlue = [System.Drawing.ColorTranslator]::FromHtml("#0A2943")
$electric = [System.Drawing.ColorTranslator]::FromHtml("#32A3F3")
$softBlue = [System.Drawing.ColorTranslator]::FromHtml("#6BC2F3")

$header = New-Canvas 150 57
$headerGraphics = [System.Drawing.Graphics]::FromImage($header)
$headerGraphics.SmoothingMode = [System.Drawing.Drawing2D.SmoothingMode]::AntiAlias
$headerGraphics.InterpolationMode = [System.Drawing.Drawing2D.InterpolationMode]::HighQualityBicubic
$headerGradient = [System.Drawing.Drawing2D.LinearGradientBrush]::new(
  [System.Drawing.Rectangle]::new(0, 0, 150, 57),
  $navy,
  $deepBlue,
  0
)
$headerGraphics.FillRectangle($headerGradient, 0, 0, 150, 57)
$headerGraphics.DrawImage($logo, 9, 10, 37, 37)
$headerTitle = [System.Drawing.Font]::new("Segoe UI", 12, [System.Drawing.FontStyle]::Bold)
$headerCaption = [System.Drawing.Font]::new("Segoe UI", 6.8, [System.Drawing.FontStyle]::Regular)
$headerGraphics.DrawString("RACORE", $headerTitle, (New-Brush "#F4FAFF"), 51, 8)
$headerGraphics.DrawString("LOCAL-FIRST BROWSER", $headerCaption, (New-Brush "#6BC2F3"), 52, 31)
$headerGraphics.DrawLine([System.Drawing.Pen]::new($electric, 2), 52, 45, 137, 45)
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
  $navy,
  $deepBlue,
  90
)
$sidebarGraphics.FillRectangle($sidebarGradient, 0, 0, 164, 314)
$sidebarGraphics.FillEllipse((New-Brush "#0D3B5B"), -55, 178, 180, 180)
$sidebarGraphics.DrawEllipse([System.Drawing.Pen]::new($softBlue, 1), -35, 198, 140, 140)
$sidebarGraphics.DrawImage($logo, 24, 23, 58, 58)

$titleFont = [System.Drawing.Font]::new("Segoe UI", 15, [System.Drawing.FontStyle]::Bold)
$smallFont = [System.Drawing.Font]::new("Segoe UI", 6.5, [System.Drawing.FontStyle]::Regular)
$labelFont = [System.Drawing.Font]::new("Consolas", 7, [System.Drawing.FontStyle]::Bold)
$sidebarGraphics.DrawString("RACORE", $titleFont, (New-Brush "#F4FAFF"), 22, 90)
$sidebarGraphics.DrawString("PRIVATE / LOCAL / AGENTIC", $smallFont, (New-Brush "#6BC2F3"), 23, 118)
$sidebarGraphics.DrawLine([System.Drawing.Pen]::new($electric, 2), 23, 140, 140, 140)

$features = @(
  @("01", "HAMMER 0.5B"),
  @("02", "LLAMA.CPP / CPU"),
  @("03", "RACORE MESH")
)
$featureY = 160
foreach ($feature in $features) {
  $sidebarGraphics.DrawString($feature[0], $smallFont, (New-Brush "#6BC2F3"), 23, $featureY)
  $sidebarGraphics.DrawString($feature[1], $labelFont, (New-Brush "#F4FAFF"), 48, ($featureY - 1))
  $featureY += 27
}

$sidebarGraphics.DrawString("NO CLOUD MODEL KEYS", $smallFont, (New-Brush "#91A8B9"), 23, 272)
$sidebarGraphics.DrawString("RACORE.XYZ / 2026", $smallFont, (New-Brush "#6BC2F3"), 23, 289)
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
