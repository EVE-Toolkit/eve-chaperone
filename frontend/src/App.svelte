<script>
  import {
    CheckAuth,
    OpenAuth,
    GetLocation,
    SwitchCurrentCharacter,
    GetRegisteredCharacters,
    LogOut,
    GetZkill,
  } from "../wailsjs/go/main/App";

  import { WindowReload } from "../wailsjs/runtime/runtime";

  let auth = false;
  let location;
  let characters;
  let killmails;

  setInterval(async () => {
    auth = await CheckAuth();
    characters = await getRegisteredCharacters();
    location = await GetLocation();
  }, 1000);

  function switchCharacter(e) {
    const character = e.target.id;

    location = GetLocation();

    switchCurrentCharacter(character);
  }

  async function addCharacter(e) {
    await OpenAuth();
    init();
  }

  function logOut(e) {
    LogOut();
  }

  async function getLocation() {
    return new Promise(async (resolve, reject) => {
      try {
        const l = await GetLocation();

        location = l;
        location = location;

        resolve(l);
      } catch (err) {
        console.error(err);

        alert("There was an error swapping characters: " + err);

        reject(err);
      }
    });
  }

  async function switchCurrentCharacter(characterName) {
    return new Promise(async (resolve, reject) => {
      try {
        await SwitchCurrentCharacter(characterName);

        WindowReload();

        await init();

        resolve();
      } catch (err) {
        console.error(err);

        alert("There was an error swapping characters: " + err);

        reject(err);
      }
    });
  }

  async function getKillMails(systemId) {
    return new Promise(async (resolve, reject) => {
      try {
        const kills = await GetZkill(systemId);

        if (!killmails) {
          setInterval(async () => {
            killmails = await GetZkill(systemId);
            killmails = killmails;

            console.log(killmails);
          }, 10000);
        }

        killmails = kills;
        killmails = killmails;

        console.log(kills);

        resolve(kills);
      } catch (err) {
        console.error(err);

        alert("There was an error getting the killmails. Please try again");

        reject(err);
      }
    });
  }

  async function getRegisteredCharacters() {
    return new Promise(async (resolve, reject) => {
      try {
        const chars = await GetRegisteredCharacters();
        characters = chars;
        characters = characters;

        console.log(chars);

        resolve(chars);
      } catch (err) {
        console.error(err);

        alert("There was an error checking auth. Please try again");

        reject(err);
      }
    });
  }

  async function getAuth() {
    return new Promise(async (resolve, reject) => {
      try {
        const authed = await CheckAuth();
        auth = auth;

        resolve(authed);
      } catch (err) {
        console.error(err);

        alert("There was an error checking auth. Please try again");

        reject(err);
      }
    });
  }

  async function init() {
    const location = await getLocation();
    getKillMails(location.solar_system_id);
  }

  // init();
</script>

<main>
  <div class="navbar bg-base-100 sticky top-0 text-black">
    <div class="navbar-start">
      <div class="dropdown">
        <div tabindex="0" role="button" class="btn btn-ghost btn-circle">
          <svg
            xmlns="http://www.w3.org/2000/svg"
            class="h-5 w-5"
            fill="none"
            viewBox="0 0 24 24"
            stroke="currentColor"
            ><path
              stroke-linecap="round"
              stroke-linejoin="round"
              stroke-width="2"
              d="M4 6h16M4 12h16M4 18h7"
            /></svg
          >
        </div>
        <ul
          tabindex="0"
          class="menu menu-sm dropdown-content mt-3 z-[1] p-2 shadow bg-base-100 rounded-box w-52"
        >
          <li><a on:click={logOut}>Log out of all accounts</a></li>
          {#if location}
            <li>
              <a href="https://anoik.is/systems/{location.name}" target="_blank"
                >Anoik.is</a
              >
            </li>
            <li>
              <a
                href="https://evemaps.dotlan.net/system/{location.name}"
                target="_blank">Dotlan</a
              >
            </li>
          {/if}
        </ul>
      </div>
    </div>
    <div class="navbar-center gap-1">
      {#await getRegisteredCharacters() then _}
        {#each characters as character}
          <button
            class="btn btn-primary w-2 text-center text-ellipsis overflow-hidden"
            id={character.name}
            on:click={switchCharacter}>{character.name}</button
          >
        {/each}
      {:catch error}
        <h1 class="text-xl text-black">Error loading auth: {error}</h1>
      {/await}
      <button class="btn btn-secondary w-2" on:click={addCharacter}>+</button>
    </div>

    <div class="navbar-end">
      {#await getAuth() then _}
        {#if auth}
          {#if location}
            <h1 class="text-xs font-bold text-black ml-2">
              {location.name}
            </h1>
          {/if}
        {:else}
          <h1 class="text-xsm font-bold text-black ml-2">...</h1>
        {/if}
      {:catch error}
        <h1 class="text-xl text-black">
          There was an error loading auth: {error}
        </h1>
      {/await}
    </div>
  </div>

  <div class="flex min-h-screen">
    <div class="m-auto">
      {#await getAuth()}
        <h1 class="text-xl text-black">Loading auth...</h1>
      {:then _}
        {#if auth && location}
          {#await GetLocation() then _}
            <h1 class="text-xl text-black">
              {#if killmails}
                <ul>
                  {#each killmails as killmail}
                    <li>
                      <a
                        class="btn btn-accent text-xsm mb-3"
                        href="https://zkillboard.com/kill/{killmail.killmailId}"
                        target="_blank"
                      >
                        <div>
                          <p>
                            {killmail.victim.ship_type_id} - {killmail.victim
                              .character_id} - {killmail.attackers.length} attackers
                          </p>
                          <div class="badge badge-sm">
                            {killmail.killmail_time}
                          </div>
                        </div></a
                      >
                    </li>
                  {/each}
                </ul>
              {:else}
                <h1 class="text-xl">Loading killmails...</h1>
              {/if}
            </h1>
          {:catch error}
            <h1 class="text-xl text-black">Error fetching location: {error}</h1>
          {/await}
        {:else}
          <h1 class="text-xl text-black">
            <h1 class="text-xl text-black mb-2">Add Character</h1>
            <button class="btn btn-primary" on:click={addCharacter}>+</button>
          </h1>
        {/if}
      {:catch error}
        <h1 class="text-xl text-black">Error loading auth: {error}</h1>
      {/await}
    </div>
  </div>
</main>
