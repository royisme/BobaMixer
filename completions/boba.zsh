#compdef boba

_boba() {
    local -a commands
    commands=(
        'ls:List profiles or adapters'
        'use:Switch to a profile'
        'stats:Display usage statistics'
        'edit:Edit configuration files'
        'doctor:Run diagnostics'
        'budget:Manage budgets'
        'hooks:Manage git hooks'
        'action:View and apply suggestions'
        'report:Generate usage reports'
        'release:Release management'
        'route:Test routing rules'
    )

    local -a ls_opts
    ls_opts=('--profiles:List all profiles')

    local -a stats_opts
    stats_opts=(
        '--today:Show today stats'
        '--7d:Show 7 day stats'
        '--30d:Show 30 day stats'
        '--by-profile:Breakdown by profile'
    )

    local -a edit_targets
    edit_targets=('profiles' 'routes' 'pricing' 'secrets')

    local -a hooks_cmds
    hooks_cmds=('install' 'remove' 'track')

    local -a route_cmds
    route_cmds=('test:Test routing rules')

    _arguments -C \
        '1: :->command' \
        '*:: :->args'

    case $state in
        command)
            _describe -t commands 'boba commands' commands
            ;;
        args)
            case $words[1] in
                ls)
                    _arguments -s -S \
                        '--profiles[List all profiles]'
                    ;;
                stats)
                    _arguments -s -S \
                        '--today[Show today stats]' \
                        '--7d[Show 7 day stats]' \
                        '--30d[Show 30 day stats]' \
                        '--by-profile[Breakdown by profile]'
                    ;;
                edit)
                    _arguments '1: :(profiles routes pricing secrets)'
                    ;;
                hooks)
                    _arguments '1: :(install remove track)'
                    ;;
                route)
                    _arguments '1: :(test)'
                    ;;
            esac
            ;;
    esac
}

_boba "$@"
