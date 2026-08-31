abstract interface class AuthSession {
  Future<String?> getIdToken();

  Future<void> signOut();
}
