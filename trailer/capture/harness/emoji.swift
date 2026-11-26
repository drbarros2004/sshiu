// Desenha emojis com a Apple Color Emoji em PNGs (emoji/<codepoints>.png), para o harness colar por cima.
import AppKit
let size: CGFloat = 96
try? FileManager.default.createDirectory(atPath: "emoji", withIntermediateDirectories: true)
for e in CommandLine.arguments.dropFirst() {
    let img = NSImage(size: NSSize(width: size, height: size))
    img.lockFocus()
    let s = NSAttributedString(string: e, attributes: [.font: NSFont(name: "Apple Color Emoji", size: size * 0.8)!])
    let b = s.size()
    s.draw(at: NSPoint(x: (size - b.width) / 2, y: (size - b.height) / 2))
    img.unlockFocus()
    let rep = NSBitmapImageRep(data: img.tiffRepresentation!)!
    let name = e.unicodeScalars.map { String($0.value, radix: 16) }.joined(separator: "-")
    try! rep.representation(using: .png, properties: [:])!.write(to: URL(fileURLWithPath: "emoji/\(name).png"))
}
