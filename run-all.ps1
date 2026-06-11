$projects = @(
    @{ Name = "Signal Room"; Path = "$PSScriptRoot\chat"; URL = "http://localhost:8081" },
    @{ Name = "Keyline"; Path = "$PSScriptRoot\auth-api"; URL = "http://localhost:8082" },
    @{ Name = "Orbit"; Path = "$PSScriptRoot\crawler"; URL = "http://localhost:8083" }
)

foreach ($project in $projects) {
    Start-Process -FilePath "go" -ArgumentList "run", "." -WorkingDirectory $project.Path -WindowStyle Hidden
    Write-Host "Started $($project.Name) at $($project.URL)"
}

Write-Host "All Go projects are running. Open portfolio/index.html to browse the collection."
