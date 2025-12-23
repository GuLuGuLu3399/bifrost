use tokio::signal;

/// Unified shutdown signal for graceful server stop.
/// Listens for Ctrl+C on all platforms and SIGTERM on Unix.
pub async fn shutdown_signal() {
    // Ctrl+C
    let ctrl_c = async {
        let _ = signal::ctrl_c().await;
    };

    // SIGTERM (Unix only); on non-Unix this future resolves immediately
    #[cfg(unix)]
    let term = async {
        use tokio::signal::unix::{signal, SignalKind};
        if let Ok(mut sig) = signal(SignalKind::terminate()) {
            sig.recv().await;
        }
    };

    #[cfg(not(unix))]
    let term = async { () };

    tokio::select! {
        _ = ctrl_c => {},
        _ = term => {},
    }
    tracing::info!("shutdown signal received");
}
