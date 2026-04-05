# Elixir Code Conventions

Recurring patterns found across 5 Phoenix/Elixir projects.

## Chained `assign_new` pipelines for derived state in components

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Function components and live components build up derived assigns via chained `assign_new/3` calls. Each step can reference assigns computed by previous steps, creating a pipeline that lazily resolves entity data, computes flags, and derives display values — all from a minimal set of input attrs. Three calling conventions coexist: `fn %{key: val}` destructuring, the `m()` macro shorthand, and `fn _ ->` for defaults.

**Example — chained entity resolution + derived flags** (from `chatty/lib/my_app_be/messages_live/20_message_card_c.html.ex`):
```elixir
  def message_card(assigns) do
    assigns =
      assigns
      |> assign_new(:message, fn %{message_id: id} ->
        MyApp.Message.resolve!(id, load: [:sending_state])
      end)
      |> assign_new(:is_email?, fn m(message) ->
        MyApp.MavuTags.resolve_uuid0("email") in List.wrap(message.tags)
      end)
      |> assign_new(:is_whatsapp?, fn m(message) ->
        MyApp.MavuTags.resolve_uuid0("whatsapp") in List.wrap(message.tags)
      end)
      |> assign_new(:is_email_stub?, fn m(message, is_email?) ->
        is_email? and email_stub?(message.content)
      end)
```

**Example — chained resolution with `%{}` destructuring** (from `wzz-ex/lib/my_app_be/filmvideo_live/filmvideo_edit_c.html.ex`):
```elixir
  def videopart_square(assigns) do
    assigns =
      assigns
      |> assign_new(:status, fn %{videopart: videopart, film: film} ->
        film.video_status["#{videopart}"]
      end)
      |> assign_new(:task_token, fn %{videopart: videopart, film: film} ->
        film.video_status["#{videopart}_token"]
      end)
      |> assign_new(:status_class, fn %{status: status} ->
        get_status_class(status) <> " text-xs font-mono rounded-sm p-[0.2rem]"
      end)
```

**Example — deep multi-step derivation** (from `eselrw/lib/my_app_be/arrows/new_arrow_for_entity_c.html.ex`):
```elixir
  def new_arrow_for_entity(assigns) do
    assigns =
      assigns
      |> assign_new(:td, fn %{arrow_name: arrow_name, target_type: target_type, source: source} ->
        {arrowtype, direction} =
          ArrowType.get_arrowtype_and_direction!(arrow_name, source.__struct__, "#{target_type}")
        %{arrowtype: arrowtype, direction: direction}
      end)
      |> assign_new(:arrowtype_config, fn %{td: td} ->
        MyApp.ArrowType.get_config(td.arrowtype, td.direction, :target)
      end)
      |> assign_new(:fake_form, fn %{td: td} ->
        create_fake_form(td.arrowtype, td.direction)
      end)
      |> assign_new(:fake_field, fn %{form: form, fake_form: fake_form, td: td, arrowtype_config: arrowtype_config} ->
        new_name = form.name <> "[new_relations][#{Uuid.to_webid0({to_string(arrowtype_config.direction), td.arrowtype.id})}]"
        field = fake_form[arrowtype_config.fieldname]
        %{field | name: new_name}
      end)
      |> assign_new(:existing_arrows, fn %{source: source, td: td} ->
        source |> Ash.load!([:arrows_ab, :arrows_ba], lazy?: true)
        ...
      end)
```

**Frequency breakdown:**
| Project | `fn m()` | `fn %{}` | `fn _ ->` (defaults) | Files with assign_new |
|---------|----------|----------|----------------------|-----------------------|
| wzz-ex | 10 | 7 | 4 | 12 |
| climate | 23 | 29 | 1 | 28 |
| chatty | 31 | 2 | 3 | 21 |
| eselrw | 5 | 12 | 3 | 14 |
| filmarchiv-ex | 9 | 9 | 3 | 7 |

**Variations:**
- `chatty` strongly prefers the `m()` macro shorthand (31 vs 2 `%{}` destructurings).
- `climate` uses both styles roughly equally and has the highest total count (28 files).
- `eselrw` and `filmarchiv-ex` lean toward `%{}` destructuring.
- Default/fallback assigns (`fn _ -> []`) appear in all 5 projects for optional CSS classes and initial state.
- Chains of 2–7 steps are common; the deepest chains appear in `eselrw` arrows and `climate` metrics modules.

---

## Shared extension-point directories and delegator modules

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** All five projects reserve the same kinds of `lib/my_app_be/` extension points: `_components/`, `_controllers/`, `my_ashfilters/`, and `my_custom_inputs/`. The `my_ashfilters` directory is especially standardized: each project exposes a `MyAppBe.MyAshfilterMapper` module that delegates to the shared base mapper and leaves a clear spot for local overrides.

**Example** (from `chatty/lib/my_app_be/my_ashfilters/my_ashfilter_mapper.ex`):
```elixir
defmodule MyAppBe.MyAshfilterMapper do
  import MavuDelegator

  # local filters go here

  delegate_all(MavuBe1.Be.Ashfilters.AshfilterMapper)
end
```

**Variations:**
- `my_custom_inputs/my_custom_input_mapper.ex` exists in all 5 projects too, but some projects keep custom rendering inline while others split it into a `MyCustomInputMapperImplementation` module.
- `my_bulkedit_components/` is present in 4/5 projects (all except chatty), so it looks common but not universal.

---

## Module naming uses `Live`, `Lc`, and `C` suffixes

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Module names consistently encode UI role in the suffix: top-level screens end in `Live`, live components often end in `Lc`, and function-component-style modules often end in `C`. In 4/5 projects (wzz-ex, climate, eselrw, filmarchiv-ex), filenames also use numeric prefixes like `00_`, `10_`, and `100_` to signal rendering/order within a screen folder.

**Example** (from `climate/ui/lib/my_app_be/climate_entites_live/10_climate_entities_list_lc.html.ex`):
```elixir
defmodule MyAppBe.ClimateEntitiesListLc do
  use MyAppBe, :live_component

  alias MyAppBe.MavuListAsh

  require Ash.Query
```

**Variations:**
- chatty uses the same `Live` / `Lc` / `C` suffixes heavily, but its files are usually not numbered.
- Some entity screens use nested namespaces like `MyAppBe.ScreeningLive.ShowtimeEditLc`; others flatten names like `MyAppBe.ClimateEntitiesListLc`.

---

## Top-of-module imports and aliases are highly standardized

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Many UI modules start with the same local DSL imports and aliases. `import Shorthand, warn: false` and `import MyApp.MavuEntities.Guards, warn: false` recur across all five projects. A shorter UUID alias also appears in 4/5 (`alias MyApp.MavuEntities.UuidHelpers, as: Uuid`).

**Example** (from `climate/ui/lib/my_app_be/climate_entites_live/10_climate_entities_list_lc.html.ex`):
```elixir
  import MyApp.MavuEntities.Guards, warn: false

  alias MyApp.MavuEntities.UuidHelpers,
    as: Uuid

  import Shorthand, warn: false
```

**Variations:**
- Some files add `warn: false` to the `Uuid` alias as well.
- Simpler modules sometimes keep only `use MyAppBe, :html` or `use MyAppBe, :live_component` and skip the helper imports entirely.

---

## Top-level LiveViews follow the same mount/params/apply_action shape

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Screen entry modules usually `use MyAppBe, :live_view`, implement `mount/3`, forward route changes through `handle_params/3`, and keep screen-specific title/state logic in `apply_action/3`.

**Example** (from `wzz-ex/lib/my_app_be/films_live/films_live.ex`):
```elixir
defmodule MyAppBe.FilmsLive do
  @moduledoc false

  use MyAppBe, :live_view

  @impl true
  def mount(_params, _session, socket) do
    {
      :ok,
      socket
    }
  end

  @impl true
  def handle_params(params, _url, socket) do
    {:noreply, apply_action(socket, socket.assigns.live_action, params)}
  end
```

**Variations:**
- Some LiveViews use `mount(params, session, socket)` instead of underscore-prefixed args when they need request/session data.
- Some screens mostly render child components and keep `apply_action/3` very small; others also subscribe to buckets, PubSub, or timers in `mount/3`.

---

## Live components are small, typed modules with `attr` declarations and optional `call/1` wrappers

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Reusable UI pieces are typically `use MyAppBe, :live_component`, declare explicit `attr` inputs, and often expose a `call/1` wrapper so parents can render them with a stable API. `update/2` and `render/1` then keep state and markup local to the component.

**Example** (from `filmarchiv-ex/lib/my_app_be/tools_live/10_action_runner_lc.html.ex`):
```elixir
defmodule MyAppBe.ToolsLive.ActionRunnerLc do
  use MyAppBe, :live_component

  import MyApp.MavuEntities.Guards, warn: false

  import Shorthand, warn: false

  attr :context, :map, required: true
  attr :tick, :integer, required: true
  attr :id, :string, required: true

  def call(assigns) do
    ~H"""
    <.live_component id={@id} module={__MODULE__} {assigns} />
    """
  end
```

**Variations:**
- Detail/list/helper components often include `call/1`; edit components are more likely to be mounted directly with `module={...}` and skip the wrapper.
- Some `:html` modules also use `attr` heavily, but expose named function components instead of live components.

---

## AshPhoenix form editors share a common validate/save/tab workflow

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Edit modules usually branch `update/2` for `new` vs existing records, build an `AshPhoenix.Form`, store tab metadata in `@tabconf`, and implement the same event set: `validate`, `save`, `add_form`, `remove_form`, and often `choose_tab`.

**Example** (from `climate/ui/lib/my_app_be/docjobs_live/docjob_edit_lc.ex`):
```elixir
  def handle_event("validate", %{"form" => params}, socket) do
    form =
      AshPhoenix.Form.validate(socket.assigns.form, params, errors: true)

    {:noreply, assign(socket, form: form)}
  end

  def handle_event("save", %{"form" => params} = msg, socket) do
    options =
      if msg["next_tab"] do
        []
      else
        [params: params |> Map.put("_current_action", "submit")]
      end
```

**Variations:**
- Some editors inject `actor: assigns[:context][:current_be_user]` into `for_create/3`.
- Some submit with `before_submit: &Resource.flatten/1`; others submit the form directly.
- More complex editors preprocess params before validation, e.g. via arrow/subform helpers.

---

## List screens are built around `MavuList` / `MavuListAsh` helper pipelines

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** List-oriented components usually centralize query creation, list config, and tweak application with shared helpers. `MavuList.process_list/4` shows up in all five projects, and 4/5 also use `MavuListAsh.get_query_for_preset!/4` plus `listconf/1` and `default_tweaks/1` helpers.

**Example** (from `filmarchiv-ex/lib/my_app_be/screenings_live/screening_list_component.ex`):
```elixir
    items_query =
      MavuListAsh.get_query_for_preset!(MyApp.Screening, preset_atom, %{},
        actor: context.current_be_user
      )

    {:ok,
     socket
     |> assign(assigns)
     |> assign(
       preset: preset_atom,
       items_query: items_query,
       selected_ids: []
     )
     |> load_items()}
```

**Variations:**
- chatty still uses `MavuList.process_list/4`, but it leans less on `MavuListAsh.get_query_for_preset!/4` than the other four projects.
- Column config is usually assembled in `listconf/1`, but the exact filter and preset wiring differs by entity.

---

## Result handling is tuple-based, with UI feedback layered on top

**Found in:** 5/5 projects (wzz-ex, climate, chatty, eselrw, filmarchiv-ex)

**Description:** Mutations are generally written around tuple-returning Ash operations. Form submits branch on `{:ok, result}` / `{:error, form}`, while destructive actions often branch on `:ok`, a structured validation error, and a fallback error. Several projects pair that with `LiveToast.send_toast/2` for immediate user feedback.

**Example** (from `filmarchiv-ex/lib/my_app_be/screengroups_live/screengroup_edit_component.ex`):
```elixir
  def handle_event("delete", _, socket) do
    Ash.destroy(socket.assigns.item)
    |> case do
      :ok ->
        LiveToast.send_toast(:info, "screengroup deleted")

        {:noreply,
         socket
         |> push_navigate(to: ~p"/be/screengroups")}

      {:error, %{errors: [%{message: message}]}} ->
        LiveToast.send_toast(:error, "screengroup cannot be deleted (#{message}) ")
        {:noreply, socket}
```

**Variations:**
- Some list actions use bang functions like `Ash.destroy!()` and skip explicit error branches.
- Form-save handlers usually reassign the error form instead of toasting, especially for validation errors.

---

## Small configuration modules stay thin and declarative

**Found in:**
- 4/5 projects for `PageHTML` template modules (climate, chatty, eselrw, filmarchiv-ex)
- 2/5 projects for `Presence` modules (eselrw, filmarchiv-ex)

**Description:** When a module exists mainly for framework wiring, these projects keep it minimal: a `use ...` macro plus a small amount of declarative config. This shows up in controller HTML modules and in Phoenix Presence setup.

**Example** (from `eselrw/lib/my_app_be/_controllers/page_html.ex`):
```elixir
defmodule MyAppBe.PageHTML do
  use MyAppBe, :html

  embed_templates "page_html/*"
end
```

**Variations:**
- `MyAppBe.Presence` in eselrw and filmarchiv-ex uses the same pattern with `use Phoenix.Presence, otp_app: :my_app, pubsub_server: MyApp.PubSub`.
- wzz-ex still has `_controllers/`, but did not expose the same `PageHTML` template module under `lib/my_app_be/`.

---
