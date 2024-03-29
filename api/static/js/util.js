function getRoundedCanvas(sourceCanvas) {
  const canvas = document.createElement('canvas')
  const context = canvas.getContext('2d')
  const width = sourceCanvas.width
  const height = sourceCanvas.height

  canvas.width = width
  canvas.height = height
  context.imageSmoothingEnabled = true
  context.drawImage(sourceCanvas, 0, 0, width, height)
  context.globalCompositeOperation = 'destination-in'
  context.beginPath()
  context.arc(
    width / 2,
    height / 2,
    Math.min(width, height) / 2,
    0,
    2 * Math.PI,
    true
  )
  context.fill()
  return canvas
}

async function createFile(url, input, fname) {
  const response = await fetch(url)
  const data = await response.blob()
  if (!fname) {
    fname = 'avatar.png'
  }
  const file = new File([data], fname, { type: data.type })
  const dt = new DataTransfer()
  dt.items.add(file)
  input.files = dt.files
}
