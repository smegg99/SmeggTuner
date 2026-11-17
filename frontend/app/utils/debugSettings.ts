import { shallowRef } from 'vue'
import type { Component } from 'vue'

// The settings a debug build adds for whoever develops the app. The plugin app/debug/settings.ts fills it;
// no other build loads that plugin, so there Settings shows nothing in its place.
export const debugSettings = shallowRef<Component>()
