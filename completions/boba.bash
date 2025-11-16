# Bash completion for boba

_boba_completions() {
    local cur prev opts
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD-1]}"

    # Main commands
    local commands="ls use stats edit doctor budget hooks action report route"

    case "${prev}" in
        boba)
            COMPREPLY=( $(compgen -W "${commands}" -- ${cur}) )
            return 0
            ;;
        ls)
            COMPREPLY=( $(compgen -W "--profiles" -- ${cur}) )
            return 0
            ;;
        stats)
            COMPREPLY=( $(compgen -W "--today --7d --30d --by-profile" -- ${cur}) )
            return 0
            ;;
        edit)
            COMPREPLY=( $(compgen -W "profiles routes pricing secrets" -- ${cur}) )
            return 0
            ;;
        budget)
            COMPREPLY=( $(compgen -W "--status --set" -- ${cur}) )
            return 0
            ;;
        hooks)
            COMPREPLY=( $(compgen -W "install remove track" -- ${cur}) )
            return 0
            ;;
        action)
            COMPREPLY=( $(compgen -W "--auto --list" -- ${cur}) )
            return 0
            ;;
        report)
            COMPREPLY=( $(compgen -W "--format --output --days" -- ${cur}) )
            return 0
            ;;
        route)
            COMPREPLY=( $(compgen -W "test" -- ${cur}) )
            return 0
            ;;
    esac
}

complete -F _boba_completions boba
