# Plaesy Constitution Kit installer for nushell (Windows, Linux, macOS).
#
# Only curl is needed. From nushell:
#
#   nu -c (curl -fsSL https://raw.githubusercontent.com/plaesy/spec-kit/main/install.nu)
#
# Downloads the release binary, verifies it against SHA256SUMS from the same
# release, and runs `plaesy install`, which picks the install directory and
# prints PATH instructions. Nothing is written to env.nu.
#
# Environment:
#   PLAESY_VERSION   release to install, e.g. v0.0.1 (default: latest)
#   PLAESY_BASE_URL  directory holding the release assets; overrides the GitHub
#                    URL (for testing or a mirror)

const REPO_URL = "https://github.com/plaesy/spec-kit"

def resolve-base-url []: nothing -> string {
    if ($env.PLAESY_BASE_URL? | is-not-empty) {
        return ($env.PLAESY_BASE_URL | str trim -r -c '/')
    }
    let version = $env.PLAESY_VERSION? | default "latest"
    if $version == "latest" {
        return $"($REPO_URL)/releases/latest/download"
    }
    if ($version !~ '^v?[0-9]+\.[0-9]+\.[0-9]+$') {
        error make {msg: $"PLAESY_VERSION must look like v1.2.3 or 'latest', got: ($version)"}
    }
    let tag = if ($version | str starts-with "v") { $version } else { $"v($version)" }
    $"($REPO_URL)/releases/download/($tag)"
}

def asset-name []: nothing -> string {
    let os = $nu.os-info.name
    let arch = $nu.os-info.arch
    match [$os, $arch] {
        ["windows", _] => "plaesy-windows-amd64.exe",
        ["linux", "x86_64"] => "plaesy-linux-amd64",
        ["linux", "aarch64"] => "plaesy-linux-arm64",
        ["macos", "aarch64"] => "plaesy-darwin-arm64",
        ["macos", "x86_64"] => (error make {msg: "No release binary for Intel Macs; build from source (see README)."}),
        _ => (error make {msg: $"No release binary for ($os)/($arch)."}),
    }
}

# Fails closed: no checksum entry for the asset means no install.
def expected-hash [sums: string, asset: string]: nothing -> string {
    let matches = open --raw $sums
        | lines
        | each {|l| $l | str trim | split row -r '\s+' }
        | where {|p| ($p | length) == 2 and (($p | last | str trim -c '*') == $asset) }
    if ($matches | is-empty) {
        error make {msg: $"No checksum for ($asset) in SHA256SUMS; refusing to install."}
    }
    $matches | first | first
}

def main [] {
    let base = resolve-base-url
    let asset = asset-name
    let tmp = ($env.TEMP? | default $env.TMPDIR? | default "/tmp") | path join $"plaesy-(random uuid)"
    mkdir $tmp
    let bin = $tmp | path join (if $nu.os-info.name == "windows" { "plaesy.exe" } else { "plaesy" })
    let sums = $tmp | path join "SHA256SUMS"

    try {
        print $"Downloading ($asset) from ($base)"
        ^curl -fsSL --tlsv1.2 -o $bin $"($base)/($asset)"
        ^curl -fsSL --tlsv1.2 -o $sums $"($base)/SHA256SUMS"

        let expected = expected-hash $sums $asset
        let actual = open --raw $bin | hash sha256
        if $actual != $expected {
            error make {msg: $"Checksum mismatch for ($asset): expected ($expected), got ($actual)."}
        }
        print "Checksum verified (SHA-256)"

        if $nu.os-info.name != "windows" { ^chmod +x $bin }
        let result = ^$bin install | complete
        print $result.stdout
        print $result.stderr
        if $result.exit_code != 0 {
            error make {msg: $"plaesy install exited with code ($result.exit_code)."}
        }
    } catch {|err|
        rm -rf $tmp
        error make {msg: $err.msg}
    }
    rm -rf $tmp
}

main
