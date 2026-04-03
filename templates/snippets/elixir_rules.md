This is a web application written using the Phoenix web framework.

## CODE RELOADING RULES
- **after you modify a *.ex file:** you have to recompile it with MavuCodeReloader.do_reload_file(file_path)
it shows you if the file compiled successfully


## DEV SERVER PORT
dont guess, always consult `tidewave-cli --port`

## PROJECT EVAL

For evaluating code in the context of the Phoenix project always exclusively use the tidewave-cli via bash

Never shell out to `iex`, `mix run`, or `elixir -e` for        
 runtime inspection. These start a separate BEAM node that cannot 
 see the running application state — live processes, GenServers,  
 ETS tables, PubSub, and supervision trees are all invisible.     
 Tidewave evaluates inside the running dev server, which is the   
 only context where live data exists. 


## STATIC CODE ANALYSIS (Giulia CLI)

Use `giulia-cli` for static code analysis and knowledge graph queries. Runs against an indexed codebase — no running server needed. Use `giulia-cli list` to see all tools.

```bash
giulia-cli brief_architect --path $PWD              # project overview + health
giulia-cli index_module_details --module MyApp.Foo --path $PWD  # module introspection
giulia-cli knowledge_dependents --module MyApp.Foo --path $PWD  # who depends on this?
giulia-cli knowledge_impact --module MyApp.Foo --path $PWD --depth 2  # blast radius
giulia-cli search_semantic --concept "edge handling" --path $PWD     # semantic search
```

Use Giulia for structure & dependencies, Tidewave for runtime state & live evaluation.


## ENVIRONMENT VARIABLES RULES

You can get set the current environment-variables by running this before any command:
`. mvp -q` sets the environment-variables for this project quietly
`. mvp -i` displays the current environment-variables for this project in the terminal

## Project guidelines

- Use the already included and available `:req` (`Req`) library for HTTP requests, **avoid** `:httpoison`, `:tesla`, and `:httpc`. Req is included by default and is the preferred HTTP client for Phoenix apps

### Phoenix v1.8 guidelines

- **Always** begin your LiveView templates with `<Layouts.app flash={@flash} ...>` which wraps all inner content
- The `MyAppWeb.Layouts` module is aliased in the `my_app_web.ex` file, so you can use it without needing to alias it again
- Anytime you run into errors with no `current_scope` assign:
  - You failed to follow the Authenticated Routes guidelines, or you failed to pass `current_scope` to `<Layouts.app>`
  - **Always** fix the `current_scope` error by moving your routes to the proper `live_session` and ensure you pass `current_scope` as needed
- Phoenix v1.8 moved the `<.flash_group>` component to the `Layouts` module. You are **forbidden** from calling `<.flash_group>` outside of the `layouts.ex` module
- Out of the box, `core_components.ex` imports an `<.icon name="hero-x-mark" class="w-5 h-5"/>` component for for hero icons. **Always** use the `<.icon>` component for icons, **never** use `Heroicons` modules or similar
- **Always** use the imported `<.input>` component for form inputs from `core_components.ex` when available. `<.input>` is imported and using it will will save steps and prevent errors
- If you override the default input classes (`<.input class="px-2 py-1 rounded-lg myclass">)`) class with your own values, no default classes are inherited, so your
custom classes must fully style the input

### JS and CSS guidelines

- **Use Tailwind4 CSS classes and custom CSS rules** to create polished, responsive, and visually stunning interfaces.

- Out of the box **only the app.js and app.css bundles are supported**
  - You cannot reference an external vendor'd script `src` or link `href` in the layouts
  - You must import the vendor deps into app.js and app.css to use them
  - **Never write inline <script>custom js</script> tags within templates**

### UI/UX & design guidelines

- **Produce world-class UI designs** with a focus on usability, aesthetics, and modern design principles
- Implement **subtle micro-interactions** (e.g., button hover effects, and smooth transitions)
- Ensure **clean typography, spacing, and layout balance** for a refined, premium look
- Focus on **delightful details** like hover effects, loading states, and smooth page transitions

