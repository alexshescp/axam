# Homebrew formula for Axam Studio (prebuilt release binaries).
# Publish as: alexshescp/homebrew-axam → Formula/axam.rb
# Install:    brew install alexshescp/axam/axam
class Axam < Formula
  desc "Axam Studio — modeless terminal text editor"
  homepage "https://github.com/alexshescp/axam"
  version "1.1.0"
  license "GPL-3.0-or-later"

  livecheck do
    url :stable
    strategy :github_latest
  end

  on_macos do
    on_arm do
      url "https://github.com/alexshescp/axam/releases/download/v#{version}/axam_darwin_arm64"
      sha256 "3166ede6d83b0f55bb0f03a0d41447f574a22b2f65d4968fc909c7148448bd97"
    end
    on_intel do
      url "https://github.com/alexshescp/axam/releases/download/v#{version}/axam_darwin_amd64"
      sha256 "38836ac1faf49f27e1b78a5eb11cb73af4a23a8446aabdb5f1c5df309dbe3661"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/alexshescp/axam/releases/download/v#{version}/axam_linux_arm64"
      sha256 "13c2e7b85c34d3b9ab9afaaace437658376e529e1fd0b65b9ef137d2149b4e4c"
    end
    on_intel do
      url "https://github.com/alexshescp/axam/releases/download/v#{version}/axam_linux_amd64"
      sha256 "e0bd5a9e0329eed5cbe4cc7f7389aa3715986d7d4b0d272afa044a98ee448077"
    end
  end

  def install
    bin.install Dir["axam_*"].reject { |f| f.end_with?(".deb", ".tar.gz") }.first => "axam"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/axam version")
    assert_match "Axam Studio", shell_output("#{bin}/axam version")
  end
end
