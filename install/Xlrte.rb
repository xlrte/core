# class Xlrte < Formula
#   full_names "xlrte"
#   desc "This is the core CLI of xlrte - DevOps with more dev and less ops."
#   homepage "https://xlrte.dev"
#   url "https://github.com/xlrte/core/archive/refs/tags/v0.1.tar.gz"
#   sha256 "b2f8e8357f68e12a6d82178ff78fe81609478f6c317509ef51494d3fb4adfbab"
#   license "Apache-2.0"


class Xlrte < Formula
  desc "Xlrte"
  homepage "https://xlrte.dev"
  version "0.1"
  license "Apache-2.0"

  if OS.mac?
    url "https://github.com/xlrte/core/releases/download/v0.1/xlrte-macos.x86.arm64"
    sha256 "1f57ece0b91f4ab4bf38824ca2956f6bbeb67c4535760acc27b0fbe0ec8950e8"
  end

  if OS.linux? && Hardware::CPU.intel?
    url "https://github.com/xlrte/core/releases/download/v0.1/xlrte-linux.x86"
    sha256 "251962f16699ad882a3000175fee5eb97d23eabb888ef81f1455d113775eb2b2"
  end

  # conflicts_with "xlrte"

  def install
    bin.install "xlrte"
  end

  # test do
  #   system "#{bin}/xlrte version"
  # end
end