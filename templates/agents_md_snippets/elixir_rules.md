## CODE RELOADING RULES
if you change a file you can recompile with MavuCodeReloader.do_reload_file(file_path)

## TIDEWAVE MCP RULES
if tidewave mcp is unreachable:
check its status with iex "MavuHealth.TidewaveHealth.report()",
and  ask me to restart it, dont use iex instead of tidewave for
anything other than the healthcheck above

## ENVIRONMENT VARIABLES RULES

You can get set the current environment-variables by running this before any command:
`. mvp` sets the environment-variables for this project
`. mvp -i` shows the current environment-variables for this project
