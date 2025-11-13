# Fish completion for boba

# Main commands
complete -c boba -f -n '__fish_use_subcommand' -a ls -d 'List profiles or adapters'
complete -c boba -f -n '__fish_use_subcommand' -a use -d 'Switch to a profile'
complete -c boba -f -n '__fish_use_subcommand' -a stats -d 'Display usage statistics'
complete -c boba -f -n '__fish_use_subcommand' -a edit -d 'Edit configuration files'
complete -c boba -f -n '__fish_use_subcommand' -a doctor -d 'Run diagnostics'
complete -c boba -f -n '__fish_use_subcommand' -a budget -d 'Manage budgets'
complete -c boba -f -n '__fish_use_subcommand' -a hooks -d 'Manage git hooks'
complete -c boba -f -n '__fish_use_subcommand' -a action -d 'View and apply suggestions'
complete -c boba -f -n '__fish_use_subcommand' -a report -d 'Generate usage reports'
complete -c boba -f -n '__fish_use_subcommand' -a release -d 'Release management'
complete -c boba -f -n '__fish_use_subcommand' -a route -d 'Test routing rules'

# ls subcommand
complete -c boba -f -n '__fish_seen_subcommand_from ls' -l profiles -d 'List all profiles'

# stats subcommand
complete -c boba -f -n '__fish_seen_subcommand_from stats' -l today -d 'Show today stats'
complete -c boba -f -n '__fish_seen_subcommand_from stats' -l 7d -d 'Show 7 day stats'
complete -c boba -f -n '__fish_seen_subcommand_from stats' -l 30d -d 'Show 30 day stats'
complete -c boba -f -n '__fish_seen_subcommand_from stats' -l by-profile -d 'Breakdown by profile'

# edit subcommand
complete -c boba -f -n '__fish_seen_subcommand_from edit' -a 'profiles routes pricing secrets'

# hooks subcommand
complete -c boba -f -n '__fish_seen_subcommand_from hooks' -a 'install remove track'

# route subcommand
complete -c boba -f -n '__fish_seen_subcommand_from route' -a 'test' -d 'Test routing rules'

# report subcommand
complete -c boba -f -n '__fish_seen_subcommand_from report' -l format -d 'Output format'
complete -c boba -f -n '__fish_seen_subcommand_from report' -l output -d 'Output file'
complete -c boba -f -n '__fish_seen_subcommand_from report' -l days -d 'Number of days'
