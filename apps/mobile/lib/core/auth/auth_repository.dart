import 'auth_gateway.dart';
import 'auth_session.dart';
import 'auth_user.dart';

final class AuthRepository implements AuthSession {
  AuthRepository(this._gateway);

  final AuthGateway _gateway;

  Stream<AuthUser?> authStateChanges() => _gateway.authStateChanges();

  Future<void> signInWithGoogle() => _gateway.signInWithGoogle();

  @override
  Future<String?> getIdToken() => _gateway.getIdToken();

  @override
  Future<void> signOut() => _gateway.signOut();
}
