cask "schnappit" do
  version "1.0.0"
  sha256 :no_check

  url "https://github.com/owenrumney/schnappit/releases/download/v#{version}/Schnappit_#{version}_macOS.dmg"
  name "Schnappit"
  desc "Lightweight screenshot capture and annotation tool"
  homepage "https://github.com/owenrumney/schnappit"

  app "Schnappit.app"

  postflight do
    # Remove quarantine attribute to avoid Gatekeeper issues
    system_command "/usr/bin/xattr",
                   args: ["-cr", "#{appdir}/Schnappit.app"],
                   sudo: false
  end

  zap trash: [
    "~/.config/schnappit",
  ]
end
