import 'auth_user.dart';

abstract interface class AuthGateway {
  Stream<AuthUser?> authStateChanges();

  Future<String?> getIdToken();

  Future<void> signInWithGoogle();

  Future<void> signOut();
}
