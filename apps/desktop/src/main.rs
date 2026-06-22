fn main() {
    tauri::Builder::<tauri::Wry>::default()
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
