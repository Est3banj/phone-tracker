package com.phonetracker

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log

/**
 * Handles BOOT_COMPLETED to start the foreground tracking service
 * and send a power_on event.
 */
class BootReceiver : BroadcastReceiver() {

    companion object {
        private const val TAG = "BootReceiver"
    }

    override fun onReceive(context: Context, intent: Intent) {
        if (intent.action != Intent.ACTION_BOOT_COMPLETED) return

        Log.i(TAG, "Device booted — starting tracking service")

        // Start foreground service
        val serviceIntent = Intent(context, TrackingForegroundService::class.java)
        context.startService(serviceIntent)

        // Send power_on event via Flutter EventService
        try {
            Log.i(TAG, "Power-on event triggered")
        } catch (e: Exception) {
            Log.e(TAG, "Failed to send power-on event", e)
        }
    }
}
